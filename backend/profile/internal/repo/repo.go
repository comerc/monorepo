package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
)

// Repo хранит профили в PostgreSQL.
type Repo struct {
	pool *pgxpool.Pool
}

// New создаёт PostgreSQL-репозиторий профилей.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Start подготавливает минимальную схему хранения профилей.
func (r *Repo) Start(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS profiles (
			user_id text PRIMARY KEY,
			nickname text UNIQUE,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now()
		)
	`)
	return err
}

// GetByUserID возвращает профиль пользователя.
func (r *Repo) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	row := r.pool.QueryRow(ctx, `SELECT user_id, nickname FROM profiles WHERE user_id = $1`, userID)

	var profile domain.Profile
	if err := row.Scan(&profile.UserID, &profile.Nickname); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

// UpsertNickname сохраняет nickname пользователя.
func (r *Repo) UpsertNickname(ctx context.Context, userID string, nickname string) (*domain.Profile, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO profiles (user_id, nickname)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET nickname = EXCLUDED.nickname, updated_at = now()
		RETURNING user_id, nickname
	`, userID, nickname)

	var profile domain.Profile
	if err := row.Scan(&profile.UserID, &profile.Nickname); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrNicknameTaken
		}
		return nil, err
	}
	return &profile, nil
}
