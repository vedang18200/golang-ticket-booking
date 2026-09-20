package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type User struct {
	ID    string
	Email string
	Role  string
}

type Service struct {
	pool       *pgxpool.Pool
	jwt        *Manager
	bcryptCost int
}

func NewService(pool *pgxpool.Pool, jwt *Manager, bcryptCost int) *Service {
	return &Service{pool: pool, jwt: jwt, bcryptCost: bcryptCost}
}

func normalize(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (s *Service) Register(ctx context.Context, email, password string) (*User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return nil, "", err
	}

	u := &User{Email: normalize(email)}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id::text, role`,
		u.Email, string(hash),
	).Scan(&u.ID, &u.Role)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return nil, "", ErrEmailTaken
	}
	if err != nil {
		return nil, "", err
	}

	token, err := s.jwt.Issue(u.ID, u.Role)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, string, error) {
	u := &User{Email: normalize(email)}
	var hash string
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, password_hash, role FROM users WHERE email = $1`, u.Email,
	).Scan(&u.ID, &hash, &u.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.jwt.Issue(u.ID, u.Role)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}
