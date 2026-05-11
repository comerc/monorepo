package auth

import (
	"context"

	authpb "github.com/pure-golang/monorepo/backend/auth/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
)

// Client вызывает auth-сервис по gRPC.
type Client struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

// New создаёт gRPC-клиент auth-сервиса.
func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

// Close закрывает gRPC-соединение.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ValidateToken проверяет JWT-токен.
func (c *Client) ValidateToken(ctx context.Context, token string) (*domain.AuthUser, error) {
	resp, err := c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return &domain.AuthUser{UserID: resp.GetUserId(), Email: resp.GetEmail()}, nil
}
