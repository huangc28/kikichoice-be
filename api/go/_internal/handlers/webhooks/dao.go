package webhooks

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/huangc28/kikichoice-be/api/go/_internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// UserDAO handles database operations for webhook user management
type UserDAO struct {
	// queries *db.Queries
	db     db.Conn
	logger *zap.SugaredLogger
}

// UserDAOParams defines dependencies for UserDAO
type UserDAOParams struct {
	fx.In

	DB     db.Conn
	Logger *zap.SugaredLogger
}

// NewUserDAO creates a new UserDAO instance
func NewUserDAO(p UserDAOParams) *UserDAO {
	return &UserDAO{
		db:     p.DB,
		logger: p.Logger,
	}
}

// CreateUserFromClerk creates a user from Clerk webhook data
func (dao *UserDAO) CreateUserFromClerk(ctx context.Context, clerkUser ClerkUser) (*db.User, error) {
	// Check if user already exists using raw SQL
	checkUserSQL := `
		SELECT id, name, email, created_at, updated_at, deleted_at, auth_provider, auth_provider_id
		FROM users
		WHERE auth_provider_id = $1 AND auth_provider = $2
	`

	var existingUser db.User
	err := dao.db.GetContext(ctx, &existingUser, checkUserSQL, clerkUser.ID, "clerk")

	if err == nil {
		// User already exists
		dao.logger.Infow("User already exists", "clerk_id", clerkUser.ID, "user_id", existingUser.ID)
		return &existingUser, nil
	}

	if err != sql.ErrNoRows {
		// Some other error occurred
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}

	// User doesn't exist, create new one
	name := clerkUser.GetFullName()
	email := clerkUser.GetPrimaryEmail()

	// Convert email string pointer to pgtype.Text
	var sqlEmail pgtype.Text
	if email != nil {
		sqlEmail = pgtype.Text{String: *email, Valid: true}
	}

	// Create user using raw SQL
	createUserSQL := `
		INSERT INTO users (name, email, auth_provider, auth_provider_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, created_at, updated_at, deleted_at, auth_provider, auth_provider_id
	`

	var user db.User
	err = dao.db.GetContext(ctx, &user, createUserSQL,
		name,
		sqlEmail,
		"clerk",
		clerkUser.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	dao.logger.Infow("Created new user from Clerk",
		"clerk_id", clerkUser.ID,
		"user_id", user.ID,
		"name", name,
		"email", email)

	return &user, nil
}
