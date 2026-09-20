package graph

import (
	"context"
	"errors"
	"time"

	"ticket-booking/graph/generated"
	"ticket-booking/graph/model"
	"ticket-booking/internal/auth"
	"ticket-booking/internal/booking"
)

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, email string, password string) (*model.AuthPayload, error) {
	if len(password) < 8 {
		return nil, codedErr("BAD_INPUT", "password must be at least 8 characters")
	}
	u, token, err := r.Auth.Register(ctx, email, password)
	if errors.Is(err, auth.ErrEmailTaken) {
		return nil, codedErr("EMAIL_TAKEN", err.Error())
	}
	if err != nil {
		return nil, internalErr(err)
	}
	return authPayload(u, token), nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*model.AuthPayload, error) {
	u, token, err := r.Auth.Login(ctx, email, password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return nil, codedErr("UNAUTHENTICATED", err.Error())
	}
	if err != nil {
		return nil, internalErr(err)
	}
	return authPayload(u, token), nil
}

// CreateEvent is the resolver for the createEvent field.
func (r *mutationResolver) CreateEvent(ctx context.Context, title string, venueID string, startsAt time.Time) (*model.Event, error) {
	return nil, codedErr("NOT_IMPLEMENTED", "createEvent not implemented yet")
}

// CreateBooking is the resolver for the createBooking field.
func (r *mutationResolver) CreateBooking(ctx context.Context, eventID string, seatID string) (*model.Booking, error) {
	id, ok := auth.FromContext(ctx)
	if !ok { // @hasRole already guarantees this; belt and braces
		return nil, codedErr("UNAUTHENTICATED", "login required")
	}

	b, err := r.Bookings.CreateNaive(ctx, id.UserID, eventID, seatID)
	if errors.Is(err, booking.ErrSeatTaken) {
		return nil, codedErr("SEAT_TAKEN", "seat already taken")
	}
	if err != nil {
		return nil, internalErr(err)
	}

	return &model.Booking{
		ID:        b.ID,
		Event:     &model.Event{ID: b.EventID},
		Seat:      &model.Seat{ID: b.SeatID},
		Status:    model.BookingStatus(b.Status),
		CreatedAt: b.CreatedAt,
	}, nil
}

// CancelBooking is the resolver for the cancelBooking field.
func (r *mutationResolver) CancelBooking(ctx context.Context, id string) (*model.Booking, error) {
	return nil, codedErr("NOT_IMPLEMENTED", "cancelBooking not implemented yet")
}

// Events is the resolver for the events field.
func (r *queryResolver) Events(ctx context.Context) ([]*model.Event, error) {
	return nil, codedErr("NOT_IMPLEMENTED", "events not implemented yet")
}

// Event is the resolver for the event field.
func (r *queryResolver) Event(ctx context.Context, id string) (*model.Event, error) {
	return nil, codedErr("NOT_IMPLEMENTED", "event not implemented yet")
}

// MyBookings is the resolver for the myBookings field.
func (r *queryResolver) MyBookings(ctx context.Context) ([]*model.Booking, error) {
	return nil, codedErr("NOT_IMPLEMENTED", "myBookings not implemented yet")
}

// SeatAvailability is the resolver for the seatAvailability field.
func (r *subscriptionResolver) SeatAvailability(ctx context.Context, eventID string) (<-chan *model.SeatUpdate, error) {
	return nil, codedErr("NOT_IMPLEMENTED", "subscriptions not implemented yet")
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
