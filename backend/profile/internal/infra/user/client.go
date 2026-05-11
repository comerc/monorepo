package user

import (
	"context"

	userpb "github.com/pure-golang/monorepo/backend/user/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
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

// GetUser возвращает пользователя по идентификатору.
func (c *Client) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	resp, err := c.client.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return &domain.User{UserID: resp.GetUserId(), Email: resp.GetEmail()}, nil
}
