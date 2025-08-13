/*
 * Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"highperf-api/internal/database"
	"highperf-api/internal/logger"
	"highperf-api/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*models.User, int, error)
}

type userRepository struct {
	db     *database.DB
	logger *logger.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB, log *logger.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: log,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.logger.WithError(err).Error("failed to create user",
			"email", user.Email,
		)
		
		// Check for unique constraint violation (PostgreSQL error code 23505)
		if isUniqueViolation(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("user created",
		"user_id", user.ID,
		"email", user.Email,
	)

	return user, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.WithError(err).Error("failed to get user by ID", "user_id", id)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.WithError(err).Error("failed to get user by email", "email", email)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		UPDATE users
		SET email = $2, name = $3, password_hash = $4, updated_at = $5
		WHERE id = $1
		RETURNING updated_at`

	user.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.UpdatedAt,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.WithError(err).Error("failed to update user", "user_id", user.ID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.Info("user updated",
		"user_id", user.ID,
		"email", user.Email,
	)

	return user, nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).Error("failed to delete user", "user_id", id)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	r.logger.Info("user deleted", "user_id", id)
	return nil
}

// List retrieves users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.WithError(err).Error("failed to count users")
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("failed to list users")
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("failed to scan user row")
			return nil, 0, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("error iterating user rows")
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, total, nil
}

// Helper function to check for unique constraint violations
func isUniqueViolation(err error) bool {
	// This is a simplified check - in production you'd want more robust error handling
	// based on your specific database driver and error types
	return err != nil && (
		// PostgreSQL unique violation
		err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" ||
		// Generic unique constraint check
		err.Error() == "UNIQUE constraint failed")
}