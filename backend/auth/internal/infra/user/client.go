package user

import (
	"context"

	userpb "github.com/pure-golang/monorepo/backend/user/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
)

// Client вызывает user-сервис по gRPC.
type Client struct {
	conn   *grpc.ClientConn
	client userpb.UserServiceClient
}

// New создаёт gRPC-клиент user-сервиса.
func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: userpb.NewUserServiceClient(conn),
	}, nil
}

// Close закрывает gRPC-соединение.
func (c *Client) Close() error {
	return c.conn.Close()
}

// GetOrCreateByEmail возвращает или создаёт пользователя по email.
func (c *Client) GetOrCreateByEmail(ctx context.Context, email string) (*domain.User, error) {
	resp, err := c.client.GetOrCreateUser(ctx, &userpb.GetOrCreateUserRequest{Email: email})
	if err != nil {
		return nil, err
	}
	return &domain.User{ID: resp.GetUserId(), Email: resp.GetEmail()}, nil
}
