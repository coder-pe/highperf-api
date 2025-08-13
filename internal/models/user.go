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

package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email" validate:"required,email"`
	Name         string    `json:"name" db:"name" validate:"required,min=2,max=100"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserCreateRequest represents the request payload for creating a user
type UserCreateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

// UserUpdateRequest represents the request payload for updating a user
type UserUpdateRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8"`
}

// UserResponse represents the response payload for a user
type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserListResponse represents the response payload for listing users
type UserListResponse struct {
	Users      []*UserResponse `json:"users"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

// ToResponse converts a User model to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToModel converts UserCreateRequest to User model
func (req *UserCreateRequest) ToModel(passwordHash string) *User {
	now := time.Now()
	return &User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ApplyUpdates applies non-nil updates to the user
func (u *User) ApplyUpdates(req *UserUpdateRequest, passwordHash string) {
	if req.Email != nil {
		u.Email = *req.Email
	}
	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Password != nil && passwordHash != "" {
		u.PasswordHash = passwordHash
	}
	u.UpdatedAt = time.Now()
}

// Pagination represents pagination parameters
type Pagination struct {
	Page    int `json:"page" validate:"min=1"`
	PerPage int `json:"per_page" validate:"min=1,max=100"`
}

// Offset calculates the database offset
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// TotalPages calculates total pages
func (p *Pagination) TotalPages(total int) int {
	if p.PerPage == 0 {
		return 0
	}
	return (total + p.PerPage - 1) / p.PerPage
}

// DefaultPagination returns default pagination values
func DefaultPagination() *Pagination {
	return &Pagination{
		Page:    1,
		PerPage: 20,
	}
}