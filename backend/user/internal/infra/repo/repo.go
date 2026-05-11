package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pure-golang/monorepo/backend/user/internal/domain"
)

// Repo хранит пользователей в PostgreSQL.
type Repo struct {
	pool *pgxpool.Pool
}

// New создаёт PostgreSQL-репозиторий пользователей.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Start подготавливает минимальную схему хранения пользователей.
func (r *Repo) Start(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id text PRIMARY KEY,
			email text NOT NULL UNIQUE,
			created_at timestamptz NOT NULL DEFAULT now()
		)
	`)
	return err
}

// GetOrCreateByEmail возвращает существующего пользователя или создаёт нового.
func (r *Repo) GetOrCreateByEmail(ctx context.Context, email string) (*domain.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, email)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id, email
	`, id.String(), email)

	var user domain.User
	if err := row.Scan(&user.ID, &user.Email); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID возвращает пользователя по идентификатору.
func (r *Repo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, email FROM users WHERE id = $1`, id)

	var user domain.User
	if err := row.Scan(&user.ID, &user.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
