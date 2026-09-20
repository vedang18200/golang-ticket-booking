package booking

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSeatTaken = errors.New("seat already taken")

type Booking struct {
	ID        string
	UserID    string
	EventID   string
	SeatID    string
	Status    string
	CreatedAt time.Time
}

type Service struct {
	pool    *pgxpool.Pool
	holdTTL time.Duration
}

func NewService(pool *pgxpool.Pool, holdTTL time.Duration) *Service {
	return &Service{pool: pool, holdTTL: holdTTL}
}

// CreateNaive is the deliberately broken version: check, then insert.
// Two requests can both pass the check before either inserts -> double booking.
// Run the k6 stampede against THIS, save the result, then fix it.
//
// Once migration 000002 (partial unique index) is applied, the losing insert fails
// with a unique violation, which we already map to ErrSeatTaken below.
func (s *Service) CreateNaive(ctx context.Context, userID, eventID, seatID string) (*Booking, error) {
	var taken bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (
		     SELECT 1 FROM bookings
		     WHERE event_id = $1 AND seat_id = $2 AND status IN ('PENDING', 'CONFIRMED')
		 )`, eventID, seatID,
	).Scan(&taken)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrSeatTaken
	}

	b := &Booking{UserID: userID, EventID: eventID, SeatID: seatID, Status: "PENDING"}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO bookings (user_id, event_id, seat_id, status, expires_at)
		 VALUES ($1, $2, $3, 'PENDING', $4)
		 RETURNING id::text, created_at`,
		userID, eventID, seatID, time.Now().Add(s.holdTTL),
	).Scan(&b.ID, &b.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrSeatTaken
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}
