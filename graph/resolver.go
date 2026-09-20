package graph

import (
	"ticket-booking/internal/auth"
	"ticket-booking/internal/booking"
)

// Resolver holds dependencies. Keep resolvers thin: real logic lives in internal/.
type Resolver struct {
	Auth     *auth.Service
	Bookings *booking.Service
}
