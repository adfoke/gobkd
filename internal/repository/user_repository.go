package repository

import (
	"context"
	"database/sql"

	"gobkd/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpsertByExternalID(ctx context.Context, user model.User) (model.User, error) {
	const upsertQuery = `
INSERT INTO users (external_id, name, email)
VALUES (?, ?, ?)
ON CONFLICT(external_id) DO UPDATE SET
	name = excluded.name,
	email = excluded.email,
	updated_at = CURRENT_TIMESTAMP;
`

	if _, err := r.db.ExecContext(ctx, upsertQuery, user.ExternalID, user.Name, user.Email); err != nil {
		return model.User{}, err
	}

	const selectQuery = `
SELECT id, external_id, name, email, created_at, updated_at
FROM users
WHERE external_id = ?;
`

	var res model.User
	err := r.db.QueryRowContext(ctx, selectQuery, user.ExternalID).Scan(
		&res.ID,
		&res.ExternalID,
		&res.Name,
		&res.Email,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	return res, nil
}
