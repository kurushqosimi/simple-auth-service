package repositories

import (
	"context"
	"errors"
	"fullstack-simple-app/internal/models"
	"fullstack-simple-app/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

const ErrUserEmailDuplicate = `ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)"`

type UserModel struct {
	pg *postgres.Postgres
}

func NewUserRepo(db *postgres.Postgres) *UserModel {
	return &UserModel{pg: db}
}

func (u *UserModel) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, created_at`

	args := []interface{}{user.FirstName, user.LastName, user.Email, user.Password.Hash}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.pg.Pool.QueryRow(
		ctx,
		query,
		args...,
	).Scan(&user.UserID, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrDuplicateEmail
		}
		return err

	}

	return nil
}

func (u *UserModel) ActivateUser(email string) (models.User, error) {
	query := `
		UPDATE users SET activated=true
		WHERE email=$1
		RETURNING user_id, first_name, last_name, created_at, activated`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	err := u.pg.Pool.QueryRow(
		ctx,
		query,
		email,
	).Scan(&user.UserID, &user.FirstName, &user.LastName, &user.CreatedAt, &user.Activated)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (u *UserModel) GetUserIDByEmail(email string) (int64, error) {
	query := `
		SELECT user_id FROM users
		WHERE email=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	err := u.pg.Pool.QueryRow(
		ctx,
		query,
		email,
	).Scan(&user.UserID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrNotFound
		}
		return 0, err
	}

	return user.UserID, nil
}

func (u *UserModel) GetUserByEmail(email string) (models.User, error) {
	query := `
		SELECT user_id, first_name, last_name, email, password_hash, active, activated FROM users
		WHERE email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	err := u.pg.Pool.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.UserID, &user.FirstName,
		&user.LastName, &user.Email,
		&user.Password.Hash, &user.Active,
		&user.Activated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, err
	}

	return user, nil
}
