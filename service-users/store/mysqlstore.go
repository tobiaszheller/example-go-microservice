package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

// User represents user on store side.
type User struct {
	ID        string    `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Nickname  string    `db:"nickname"`
	Email     string    `db:"email"`
	Country   string    `db:"country"`
	UpdatedAt time.Time `db:"updated_at"`
}

type store struct {
	db *sqlx.DB
}

func New(db *sql.DB) *store {
	return &store{
		db: sqlx.NewDb(db, "mysql"),
	}
}

func (s *store) CreateUser(ctx context.Context, in *User) (*User, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", err)
	}
	in.ID = uuid.String()
	in.UpdatedAt = time.Now().UTC()
	res, err := s.db.NamedExecContext(ctx, queryInsertUser, in)
	if err != nil {
		if isMysqlDuplicateEntryErr(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("cannot check affected rows: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("user not created, no affected rows")
	}
	return in, err
}

func (s *store) UpdateUser(ctx context.Context, in *User) (*User, error) {
	in.UpdatedAt = time.Now().UTC()
	res, err := s.db.NamedExecContext(ctx, queryUpdateUser, in)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("cannot check affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrUserNotFound
	}
	return in, err
}

func (s *store) GetUser(ctx context.Context, id string) (*User, error) {
	var out User
	if err := s.db.GetContext(ctx, &out, querySelectUserById, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &out, nil
}

func isMysqlDuplicateEntryErr(err error) bool {
	if err == nil {
		return false
	}
	me, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return me.Number == 1062
}
