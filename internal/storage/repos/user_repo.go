package repos

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"gophermart/internal/dferrors"
)

type User struct {
	ID           int
	Login        string
	PasswordHash string
}

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int, error)
	GetPasswordHash(ctx context.Context, login string) (string, error)
	GetUserByLogin(ctx context.Context, login string) (*User, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, login, passwordHash string) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int
	err = r.db.QueryRowContext(
		ctx,
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id",
		login,
		passwordHash,
	).Scan(&id)

	if err != nil {
		if isUniqueViolation(err) {
			return 0, dferrors.ErrUserAlreadyExists
		}
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO wallets (user_id) VALUES ($1)`, id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, errors.New("wallet already exists")
		}
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresUserRepository) GetUserByLogin(
	ctx context.Context,
	login string,
) (*User, error) {
	var u User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, username, password_hash FROM users WHERE username = $1",
		login,
	).Scan(&u.ID, &u.Login, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return &u, err
}

func (r *PostgresUserRepository) GetPasswordHash(
	ctx context.Context,
	login string,
) (string, error) {
	var hash string
	err := r.db.QueryRowContext(
		ctx,
		"SELECT password_hash FROM users WHERE username = $1",
		login,
	).Scan(&hash)
	return hash, err
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code.Name() == "unique_violation"
	}
	return false
}
