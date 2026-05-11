//go:build bdd

package bdd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcrabbitmq "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const (
	postgresImage = "postgres:16-alpine"
	redisImage    = "redis:7-alpine"
	rabbitImage   = "rabbitmq:3.13-alpine"
)

var (
	liveStack    *Stack
	liveStackErr error
	stackOnce    sync.Once
)

// Stack описывает запущенный набор внешних зависимостей и сервисов.
type Stack struct {
	authHTTP    string
	profileHTTP string
	gatewayHTTP string

	postgres  *tcpostgres.PostgresContainer
	redis     *tcredis.RedisContainer
	rabbit    *tcrabbitmq.RabbitMQContainer
	processes []*serviceProcess
	tempDirs  []string
	client    *http.Client

	smtpListener net.Listener
	mailMu       sync.Mutex
	mails        []capturedMail
}

type serviceProcess struct {
	name string
	cmd  *exec.Cmd
	log  *bytes.Buffer
}

type graphQLResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type graphQLErrors struct {
	messages []string
}

func (e graphQLErrors) Error() string {
	return "graphql errors: " + strings.Join(e.messages, "; ")
}

type capturedMail struct {
	To   []string
	Body string
}

type jwtPayload struct {
	TokenID string `json:"jti"`
	UserID  string `json:"sub"`
	Email   string `json:"email"`
}

// GetOrCreateStack создаёт общий black-box stack для BDD-прогона.
func GetOrCreateStack(t *testing.T) (*Stack, error) {
	t.Helper()

	stackOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		liveStack, liveStackErr = startStack(ctx)
		if liveStackErr == nil {
			t.Cleanup(func() {
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer shutdownCancel()
				liveStack.Close(shutdownCtx)
			})
		}
	})
	return liveStack, liveStackErr
}

func startStack(ctx context.Context) (*Stack, error) {
	stack := &Stack{client: &http.Client{Timeout: 5 * time.Second}}
	if err := stack.startContainers(ctx); err != nil {
		stack.closeWithTimeout()
		return nil, err
	}
	if err := stack.startServices(ctx); err != nil {
		stack.closeWithTimeout()
		return nil, err
	}
	return stack, nil
}

func (s *Stack) closeWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Close(ctx)
}

func (s *Stack) startContainers(ctx context.Context) error {
	postgres, err := tcpostgres.Run(ctx, postgresImage,
		tcpostgres.WithDatabase("bdd"),
		tcpostgres.WithUsername("bdd"),
		tcpostgres.WithPassword("bdd"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return fmt.Errorf("start postgres: %w", err)
	}
	s.postgres = postgres
	if err := waitPostgres(ctx, postgres); err != nil {
		return err
	}

	redis, err := tcredis.Run(ctx, redisImage)
	if err != nil {
		return fmt.Errorf("start redis: %w", err)
	}
	s.redis = redis

	rabbit, err := tcrabbitmq.Run(ctx, rabbitImage)
	if err != nil {
		return fmt.Errorf("start rabbitmq: %w", err)
	}
	s.rabbit = rabbit
	if err := waitRabbitMQ(ctx, rabbit); err != nil {
		return err
	}

	return s.startSMTP()
}

func (s *Stack) startServices(ctx context.Context) error {
	postgresEnv, err := containerPostgresEnv(ctx, s.postgres)
	if err != nil {
		return err
	}
	redisEnv, err := containerRedisEnv(ctx, s.redis)
	if err != nil {
		return err
	}
	rabbitURL, err := s.rabbit.AmqpURL(ctx)
	if err != nil {
		return err
	}
	mailHost, mailPort, err := splitHostPort(s.smtpListener.Addr().String())
	if err != nil {
		return fmt.Errorf("resolve smtp listener address: %w", err)
	}

	userGRPC := freePort()
	authHTTP := freePort()
	authGRPC := freePort()
	profileHTTP := freePort()
	profileGRPC := freePort()
	gatewayHTTP := freePort()

	common := make([]string, 0, len(postgresEnv)+len(redisEnv)+4)
	common = append(common, postgresEnv...)
	common = append(common, redisEnv...)
	common = append(common,
		"ENVIRONMENT=test",
		"RABBITMQ_URL="+rabbitURL,
		"TRACING_ENDPOINT=",
		"METRICS_PORT=0",
	)

	userProc, err := s.startProcess("user", "./backend/user/cmd/server", append(common,
		"GRPC_HOST=127.0.0.1",
		"GRPC_PORT="+userGRPC,
	))
	if err != nil {
		return err
	}
	if err := waitTCP(ctx, "127.0.0.1:"+userGRPC); err != nil {
		return fmt.Errorf("user service did not become ready: %w\n%s", err, userProc.output())
	}

	authProc, err := s.startProcess("auth", "./backend/auth/cmd/server", append(common,
		"USER_GRPC_ADDR=127.0.0.1:"+userGRPC,
		"AUTH_JWT_SECRET=bdd-secret",
		"AUTH_CODE_TTL=10m",
		"SMTP_HOST="+mailHost,
		"SMTP_PORT="+mailPort,
		"SMTP_FROM=auth@example.test",
		"SMTP_TLS=false",
		"SMTP_INSECURE=true",
		"SMTP_MAX_RETRIES=1",
		"WEBSERVER_HOST=127.0.0.1",
		"WEBSERVER_PORT="+authHTTP,
		"GRPC_HOST=127.0.0.1",
		"GRPC_PORT="+authGRPC,
	))
	if err != nil {
		return err
	}
	s.authHTTP = "http://127.0.0.1:" + authHTTP + "/graphql"
	if err := s.waitGraphQL(ctx, s.authHTTP, `query { authStatus }`, "authStatus"); err != nil {
		return fmt.Errorf("auth service did not become ready: %w\n%s", err, authProc.output())
	}

	profileProc, err := s.startProcess("profile", "./backend/profile/cmd/server", append(common,
		"AUTH_GRPC_ADDR=127.0.0.1:"+authGRPC,
		"WEBSERVER_HOST=127.0.0.1",
		"WEBSERVER_PORT="+profileHTTP,
		"GRPC_HOST=127.0.0.1",
		"GRPC_PORT="+profileGRPC,
	))
	if err != nil {
		return err
	}
	s.profileHTTP = "http://127.0.0.1:" + profileHTTP + "/graphql"
	if err := s.waitGraphQL(ctx, s.profileHTTP, `query { profileStatus }`, "profileStatus"); err != nil {
		return fmt.Errorf("profile service did not become ready: %w\n%s", err, profileProc.output())
	}

	gatewayProc, err := s.startGateway(ctx, gatewayHTTP, authHTTP, profileHTTP)
	if err != nil {
		return err
	}
	s.gatewayHTTP = "http://127.0.0.1:" + gatewayHTTP + "/graphql"
	if err := s.waitGatewayGraphQL(ctx); err != nil {
		return fmt.Errorf("gateway did not become ready: %w\n%s", err, gatewayProc.output())
	}
	return nil
}

func (s *Stack) startGateway(ctx context.Context, gatewayHTTP string, authHTTP string, profileHTTP string) (*serviceProcess, error) {
	workDir, err := os.MkdirTemp("", "monorepo-bdd-gateway-*")
	if err != nil {
		return nil, fmt.Errorf("create gateway temp dir: %w", err)
	}

	routerJSON, err := os.ReadFile(filepath.Join(projectRoot(), "backend", "gateway", "router.json"))
	if err != nil {
		return nil, fmt.Errorf("read gateway router config: %w", err)
	}
	routerConfig := strings.ReplaceAll(string(routerJSON), "http://localhost:8081/graphql", "http://127.0.0.1:"+authHTTP+"/graphql")
	routerConfig = strings.ReplaceAll(routerConfig, "http://localhost:8082/graphql", "http://127.0.0.1:"+profileHTTP+"/graphql")
	routerConfigPath := filepath.Join(workDir, "router.json")
	if err := writeTempFile(routerConfigPath, []byte(routerConfig)); err != nil {
		return nil, fmt.Errorf("write gateway router config: %w", err)
	}
	s.tempDirs = append(s.tempDirs, workDir)

	config := fmt.Sprintf(`version: "1"

listen_addr: "127.0.0.1:%s"
playground_enabled: false
log_level: "info"

execution_config:
  file:
    path: "router.json"
    watch: false

headers:
  all:
    request:
      - op: "propagate"
        named: Authorization
`, gatewayHTTP)
	configPath := filepath.Join(workDir, "config.yaml")
	if err := writeTempFile(configPath, []byte(config)); err != nil {
		return nil, fmt.Errorf("write gateway runtime config: %w", err)
	}

	routerBin, err := cosmoRouterBinary(ctx)
	if err != nil {
		return nil, err
	}

	proc, err := s.startProcessInDir("gateway", workDir, routerBin, []string{"-config", configPath}, nil)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (s *Stack) startProcess(name string, pkg string, env []string) (*serviceProcess, error) {
	return s.startProcessInDir(name, projectRoot(), "go", []string{"run", pkg}, env)
}

func (s *Stack) startProcessInDir(name string, dir string, command string, args []string, env []string) (*serviceProcess, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	log := &bytes.Buffer{}
	cmd.Stdout = log
	cmd.Stderr = log

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	proc := &serviceProcess{name: name, cmd: cmd, log: log}
	s.processes = append(s.processes, proc)
	return proc, nil
}

func cosmoRouterBinary(ctx context.Context) (string, error) {
	if path := os.Getenv("COSMO_ROUTER_BIN"); path != "" {
		return path, nil
	}

	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}
	routerDir := filepath.Join(cacheRoot, "pure-golang-monorepo", "cosmo-router")
	routerBin := filepath.Join(routerDir, "router")
	if _, err := os.Stat(routerBin); err == nil {
		return routerBin, nil
	}
	if err := os.MkdirAll(routerDir, 0o755); err != nil {
		return "", fmt.Errorf("create cosmo router cache dir: %w", err)
	}

	cmd := exec.CommandContext(ctx, "npx", "wgc", "router", "download-binary", "-o", routerDir)
	cmd.Dir = filepath.Join(projectRoot(), "backend", "gateway")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("download cosmo router binary: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	if _, err := os.Stat(routerBin); err != nil {
		return "", fmt.Errorf("cosmo router binary was not downloaded: %w", err)
	}
	return routerBin, nil
}

func writeTempFile(path string, data []byte) error {
	cleanPath := filepath.Clean(path)
	file, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close temp file %s: %v\n", cleanPath, err)
		}
	}()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return nil
}

func (p *serviceProcess) output() string {
	if p == nil || p.log == nil {
		return ""
	}
	out := strings.TrimSpace(p.log.String())
	if out == "" {
		return p.name + " output is empty"
	}
	return p.name + " output:\n" + out
}

// Close останавливает процессы и контейнеры BDD stack.
func (s *Stack) Close(ctx context.Context) {
	if s.smtpListener != nil {
		if err := s.smtpListener.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close smtp listener: %v\n", err)
		}
	}
	for i := len(s.processes) - 1; i >= 0; i-- {
		proc := s.processes[i]
		if proc.cmd.Process != nil {
			if err := syscall.Kill(-proc.cmd.Process.Pid, syscall.SIGKILL); err != nil {
				fmt.Fprintf(os.Stderr, "kill %s process group: %v\n", proc.name, err)
			}
			if err := proc.cmd.Wait(); err != nil && proc.cmd.ProcessState == nil {
				fmt.Fprintf(os.Stderr, "wait %s process: %v\n", proc.name, err)
			}
		}
	}
	if s.rabbit != nil {
		if err := s.rabbit.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "terminate rabbitmq container: %v\n", err)
		}
	}
	for _, dir := range s.tempDirs {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(os.Stderr, "remove temp dir %s: %v\n", dir, err)
		}
	}
	if s.redis != nil {
		if err := s.redis.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "terminate redis container: %v\n", err)
		}
	}
	if s.postgres != nil {
		if err := s.postgres.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "terminate postgres container: %v\n", err)
		}
	}
}

// RequestEmailCode запрашивает email-код через федеративный GraphQL.
func (s *Stack) RequestEmailCode(ctx context.Context, email string) (int, error) {
	after := s.MailCount()

	var payload struct {
		RequestEmailCode struct {
			Accepted bool `json:"accepted"`
		} `json:"requestEmailCode"`
	}
	err := s.graphQL(ctx, s.gatewayHTTP, "", `mutation ($email: String!) {
		requestEmailCode(email: $email) { accepted }
	}`, map[string]any{"email": email}, &payload)
	if err != nil {
		return after, err
	}
	if !payload.RequestEmailCode.Accepted {
		return after, fmt.Errorf("requestEmailCode was not accepted")
	}
	return after, nil
}

// WaitEmailCode ждёт письмо и возвращает найденный 4-значный код.
func (s *Stack) WaitEmailCode(ctx context.Context, email string, after int) (string, error) {
	codePattern := regexp.MustCompile(`(?i)access code is (\d{4})`)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		s.mailMu.Lock()
		mails := append([]capturedMail(nil), s.mails[after:]...)
		s.mailMu.Unlock()

		for i := len(mails) - 1; i >= 0; i-- {
			msg := mails[i]
			if !mailSentTo(msg, email) {
				continue
			}
			matches := codePattern.FindStringSubmatch(msg.Body)
			if len(matches) == 2 {
				return matches[1], nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return "", fmt.Errorf("email code for %q was not received", email)
}

// MailCount возвращает количество перехваченных писем.
func (s *Stack) MailCount() int {
	s.mailMu.Lock()
	defer s.mailMu.Unlock()
	return len(s.mails)
}

// RequestEmailCodeAndWait запрашивает свежий email-код.
func (s *Stack) RequestEmailCodeAndWait(ctx context.Context, email string) (string, error) {
	after, err := s.RequestEmailCode(ctx, email)
	if err != nil {
		return "", err
	}
	return s.WaitEmailCode(ctx, email, after)
}

// LoginByEmail выполняет полный вход по email-коду.
func (s *Stack) LoginByEmail(ctx context.Context, email string) (string, error) {
	code, err := s.RequestEmailCodeAndWait(ctx, email)
	if err != nil {
		return "", err
	}
	return s.LoginWithEmailCode(ctx, email, code)
}

// ExpectJWT проверяет, что выданный токен похож на JWT с данными пользователя.
func (s *Stack) ExpectJWT(ctx context.Context, token string, email string) error {
	payload, err := decodeJWTPayload(token)
	if err != nil {
		return err
	}
	if payload.Email != email {
		return fmt.Errorf("jwt email mismatch: got %q, want %q", payload.Email, email)
	}
	if payload.UserID == "" {
		return fmt.Errorf("jwt user id is empty")
	}
	if payload.TokenID == "" {
		return fmt.Errorf("jwt token id is empty")
	}
	if _, err := s.MyNickname(ctx, token); err != nil {
		return fmt.Errorf("jwt was not accepted by federated graphql: %w", err)
	}
	return nil
}

// LoginWithEmailCode вводит email-код и возвращает JWT.
func (s *Stack) LoginWithEmailCode(ctx context.Context, email string, code string) (string, error) {
	var payload struct {
		LoginWithEmailCode struct {
			Token string `json:"token"`
		} `json:"loginWithEmailCode"`
	}
	err := s.graphQL(ctx, s.gatewayHTTP, "", `mutation ($email: String!, $code: String!) {
		loginWithEmailCode(email: $email, code: $code) { token }
	}`, map[string]any{"email": email, "code": code}, &payload)
	if err != nil {
		return "", err
	}
	return payload.LoginWithEmailCode.Token, nil
}

// Logout отзывает JWT через федеративный GraphQL.
func (s *Stack) Logout(ctx context.Context, token string) error {
	var payload struct {
		Logout struct {
			Revoked bool `json:"revoked"`
		} `json:"logout"`
	}
	err := s.graphQL(ctx, s.gatewayHTTP, token, `mutation { logout { revoked } }`, nil, &payload)
	if err != nil {
		return err
	}
	if !payload.Logout.Revoked {
		return fmt.Errorf("logout did not revoke token")
	}
	return nil
}

// ExpectTokenRejected проверяет, что auth GraphQL больше не принимает JWT.
func (s *Stack) ExpectTokenRejected(ctx context.Context, token string) error {
	var payload struct {
		Logout struct {
			Revoked bool `json:"revoked"`
		} `json:"logout"`
	}
	err := s.graphQL(ctx, s.authHTTP, token, `mutation { logout { revoked } }`, nil, &payload)
	if err == nil {
		return fmt.Errorf("revoked token was accepted by auth graphql")
	}
	var gqlErr graphQLErrors
	if !errors.As(err, &gqlErr) {
		return fmt.Errorf("expected revoked token graphql error, got %w", err)
	}
	for _, message := range gqlErr.messages {
		if strings.Contains(message, "token is revoked") {
			return nil
		}
	}
	return fmt.Errorf("expected revoked token graphql error, got %w", err)
}

// SetNickname сохраняет nickname через федеративный GraphQL.
func (s *Stack) SetNickname(ctx context.Context, token string, nickname string) error {
	var payload struct {
		SetNickname struct {
			Nickname string `json:"nickname"`
		} `json:"setNickname"`
	}
	return s.graphQL(ctx, s.gatewayHTTP, token, `mutation ($nickname: String!) {
		setNickname(nickname: $nickname) { nickname }
	}`, map[string]any{"nickname": nickname}, &payload)
}

// MyNickname читает nickname текущего пользователя через федеративный GraphQL.
func (s *Stack) MyNickname(ctx context.Context, token string) (string, error) {
	var payload struct {
		MyProfile struct {
			Nickname string `json:"nickname"`
		} `json:"myProfile"`
	}
	if err := s.graphQL(ctx, s.gatewayHTTP, token, `query { myProfile { nickname } }`, nil, &payload); err != nil {
		return "", err
	}
	return payload.MyProfile.Nickname, nil
}

func (s *Stack) waitGatewayGraphQL(ctx context.Context) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		var data struct {
			AuthStatus    string `json:"authStatus"`
			ProfileStatus string `json:"profileStatus"`
		}
		err := s.graphQL(ctx, s.gatewayHTTP, "", `query { authStatus profileStatus }`, nil, &data)
		if err == nil && data.AuthStatus == "ok" && data.ProfileStatus == "ok" {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("graphql endpoint %s did not become ready", s.gatewayHTTP)
}

func (s *Stack) waitGraphQL(ctx context.Context, url string, query string, field string) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		var data map[string]string
		if err := s.graphQL(ctx, url, "", query, nil, &data); err == nil && data[field] == "ok" {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("graphql endpoint %s did not become ready", url)
}

func (s *Stack) graphQL(ctx context.Context, url string, token string, query string, variables map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close graphql response body: %v\n", err)
		}
	}()

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql status %d", resp.StatusCode)
	}
	if len(gqlResp.Errors) > 0 {
		messages := make([]string, 0, len(gqlResp.Errors))
		for _, gqlErr := range gqlResp.Errors {
			messages = append(messages, gqlErr.Message)
		}
		return graphQLErrors{messages: messages}
	}
	if out == nil {
		return nil
	}
	data, err := json.Marshal(gqlResp.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func containerPostgresEnv(ctx context.Context, container *tcpostgres.PostgresContainer) ([]string, error) {
	endpoint, err := container.PortEndpoint(ctx, "5432/tcp", "")
	if err != nil {
		return nil, err
	}
	host, port, err := splitHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	return []string{
		"POSTGRES_USER=bdd",
		"POSTGRES_PASSWORD=bdd",
		"POSTGRES_DB=bdd",
		"POSTGRES_HOST=" + host,
		"POSTGRES_PORT=" + port,
	}, nil
}

func containerRedisEnv(ctx context.Context, container *tcredis.RedisContainer) ([]string, error) {
	connString, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}
	addr, err := connStringHostPort(connString)
	if err != nil {
		return nil, err
	}
	return []string{"REDIS_ADDR=" + addr}, nil
}

func waitRabbitMQ(ctx context.Context, container *tcrabbitmq.RabbitMQContainer) error {
	url, err := container.AmqpURL(ctx)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := amqp.Dial(url)
		if err == nil {
			if closeErr := conn.Close(); closeErr != nil {
				return fmt.Errorf("close RabbitMQ readiness connection: %w", closeErr)
			}
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("rabbitmq did not become ready")
}

func waitPostgres(ctx context.Context, container *tcpostgres.PostgresContainer) error {
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			pingErr := pool.Ping(ctx)
			pool.Close()
			if pingErr == nil {
				return nil
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("postgres did not become ready")
}

func freePort() string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("failed to allocate free port: " + err.Error())
	}
	defer func() {
		if err := listener.Close(); err != nil {
			panic("failed to close free port listener: " + err.Error())
		}
	}()
	return fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
}

func waitTCP(ctx context.Context, addr string) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			if err := conn.Close(); err != nil {
				return fmt.Errorf("close tcp probe connection: %w", err)
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
	return fmt.Errorf("tcp endpoint %s did not become ready", addr)
}

func (s *Stack) startSMTP() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start smtp listener: %w", err)
	}
	s.smtpListener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go s.handleSMTP(conn)
		}
	}()
	return nil
}

func (s *Stack) handleSMTP(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close smtp connection: %v\n", err)
		}
	}()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	writeSMTP(writer, "220 bdd smtp ready")

	var recipients []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.TrimSpace(line)
		upper := strings.ToUpper(cmd)

		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			writeSMTP(writer, "250-bdd")
			writeSMTP(writer, "250 OK")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			writeSMTP(writer, "250 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			recipients = append(recipients, parseSMTPAddress(cmd))
			writeSMTP(writer, "250 OK")
		case upper == "DATA":
			writeSMTP(writer, "354 End data with <CR><LF>.<CR><LF>")
			var body strings.Builder
			for {
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimSpace(dataLine) == "." {
					break
				}
				if _, err := body.WriteString(dataLine); err != nil {
					return
				}
			}
			s.mailMu.Lock()
			s.mails = append(s.mails, capturedMail{To: append([]string(nil), recipients...), Body: body.String()})
			s.mailMu.Unlock()
			writeSMTP(writer, "250 OK")
		case upper == "RSET":
			recipients = nil
			writeSMTP(writer, "250 OK")
		case upper == "NOOP":
			writeSMTP(writer, "250 OK")
		case upper == "QUIT":
			writeSMTP(writer, "221 Bye")
			return
		default:
			writeSMTP(writer, "250 OK")
		}
	}
}

func writeSMTP(writer *bufio.Writer, line string) {
	if _, err := writer.WriteString(line + "\r\n"); err != nil {
		return
	}
	if err := writer.Flush(); err != nil {
		return
	}
}

func parseSMTPAddress(line string) string {
	start := strings.Index(line, "<")
	end := strings.LastIndex(line, ">")
	if start >= 0 && end > start {
		return line[start+1 : end]
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func decodeJWTPayload(token string) (*jwtPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("jwt token must contain 3 parts")
	}
	payloadData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode jwt payload: %w", err)
	}
	var payload jwtPayload
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		return nil, fmt.Errorf("parse jwt payload: %w", err)
	}
	return &payload, nil
}

func mailSentTo(msg capturedMail, email string) bool {
	for _, to := range msg.To {
		if strings.EqualFold(to, email) {
			return true
		}
	}
	return false
}

func splitHostPort(addr string) (string, string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", "", err
	}
	return host, port, nil
}

func connStringHostPort(connString string) (string, error) {
	parsed, err := neturl.Parse(connString)
	if err != nil {
		return "", err
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("connection string host is empty")
	}
	return parsed.Host, nil
}

func projectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic("failed to resolve cwd: " + err.Error())
	}
	if filepath.Base(wd) == "bdd" {
		return filepath.Join(wd, "..", "..")
	}
	return wd
}
