package repository

import (
	"context"
	"database/sql"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindOrCreateUser(ctx context.Context, appleSub string, email *string, firstName *string, lastName *string) (models.User, error) {
	var user models.User

	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, first_name, last_name
		FROM users
		WHERE apple_sub = $1
	`, appleSub)
	err := row.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName)
	if err == nil {
		return user, nil
	}
	
	if err != sql.ErrNoRows {
		return user, err
	}

	row = r.db.QueryRowContext(ctx, `
		INSERT INTO users (apple_sub, email, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, first_name, last_name
	`, appleSub, email, firstName, lastName)
	err = row.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName)
	
	return user, err
}