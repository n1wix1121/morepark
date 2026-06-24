package repository

import (
	"context"
	"fmt"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail находит пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// GetByID находит пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}
		user.PasswordHash = ""
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, full_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, full_name = $2, role = $3, is_active = $4, password_hash = COALESCE(NULLIF($5, ''), password_hash)
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, query,
		user.Email,
		user.FullName,
		user.Role,
		user.IsActive,
		user.PasswordHash,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *UserRepository) CountByRole(ctx context.Context, role string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = $1`, role).Scan(&count)
	return count, err
}
