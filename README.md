# Ticket Booking

An event-driven ticket booking backend in Go and GraphQL, built around one hard problem: **thousands of people fighting for the same seat.** The goal is simple and measurable: 1,000 concurrent users, 50 seats, zero double-bookings. Seat reservation is concurrency-safe, and everything after a booking (payment, notifications, ticket generation) happens asynchronously through a message queue.

**Stack:** Go · gqlgen (GraphQL) · PostgreSQL · Redis · RabbitMQ · Docker Compose · k6

> **Status: in progress.** The roadmap at the bottom shows what's done. Load-test numbers are added as each stage lands.

## Architecture

```mermaid
flowchart LR
    C[Client] -->|GraphQL / WebSocket| API["API<br/>Go + gqlgen"]
    API -->|users, bookings| PG[(PostgreSQL)]
    API -->|seat holds, pub/sub| R[(Redis)]
    API -->|BookingCreated| MQ{{RabbitMQ}}
    MQ --> PW[Payment worker]
    MQ --> NW[Notification worker]
    MQ --> TW[Ticket / QR worker]
    PW -->|status updates| PG
```

- **API**: a thin GraphQL layer; real logic lives in `internal/`.
- **PostgreSQL**: the source of truth for who owns which seat.
- **Redis**: short-lived seat holds and pub/sub for live availability.
- **RabbitMQ + workers**: each step after a booking is its own Go binary, so a slow payment provider never blocks seat reservation.

## The concurrency problem

Two users click the same seat at the same moment. A naive "check if free, then insert" lets both through.

1. **Baseline:** the first version is deliberately naive (check-then-insert). The k6 stampede reproduces the double-bookings.
2. **Fix:** a partial unique index on `(event_id, seat_id)` for active bookings makes Postgres the final arbiter, with a Redis hold in front so most conflicts fail fast without touching the database.
3. **Proof:** the same k6 test, plus a SQL check that no seat has more than one active booking.

### Results

| Version | Users | Seats | Bookings created | Double-booked seats | p95 latency |
|---|---|---|---|---|---|
| Naive check-then-insert | 1,000 | 50 | — | — | — |
| Unique index + Redis hold | 1,000 | 50 | — | — | — |

*Numbers are filled in from real `k6` runs (`loadtest/stampede.js`).*

## Quickstart

**Prerequisites:** Go 1.22+, Docker, [golang-migrate](https://github.com/golang-migrate/migrate), [k6](https://k6.io)

```bash
git clone https://github.com/vedang18200/ticket-booking
cd ticket-booking

make up        # PostgreSQL, Redis, RabbitMQ via Docker Compose
make migrate   # apply database migrations
make seed      # 1 venue, 50 seats, 1 event
make api       # GraphQL API on http://localhost:8080 (playground at /)
```

Try it:

```graphql
mutation {
  register(email: "you@example.dev", password: "password123") {
    token
    user { id role }
  }
}
```

Send the token as `Authorization: Bearer <token>` for protected mutations such as `createBooking`.

### Configuration

| Variable | Default |
|---|---|
| `PORT` | `8080` |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/booking?sslmode=disable` |
| `REDIS_ADDR` | `localhost:6379` |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` |
| `JWT_SECRET` | `dev-secret-change-me` (**change in production**) |
| `SEAT_HOLD_TTL` | `5m` |
| `BCRYPT_COST` | `10` (use `4` for load tests) |

## Load testing

```bash
export EVENT_ID=$(docker compose exec -T postgres psql -U postgres -d booking -Atc "select id from events")
export SEAT_IDS=$(docker compose exec -T postgres psql -U postgres -d booking -Atc "select json_agg(id) from seats")
BCRYPT_COST=4 make api   # in one terminal
make load                # in another
```

The test registers 1,000 users, then all of them try to book random seats among the 50 at once. Afterwards this query must return zero rows:

```sql
SELECT seat_id, count(*) FROM bookings
WHERE status IN ('PENDING', 'CONFIRMED')
GROUP BY seat_id HAVING count(*) > 1;
```

## Project layout

```
cmd/
  api/                   GraphQL server
  payment-worker/        consumes BookingCreated -> mock payment
  notification-worker/   email / notification
  ticket-worker/         QR ticket generation
graph/                   gqlgen schema, generated code, thin resolvers
internal/
  auth/                  JWT, middleware, credentials
  booking/               reservation logic
  config/  db/           configuration, Postgres pool
migrations/              golang-migrate SQL files
loadtest/                k6 scripts
docs/DESIGN.md           design decisions and trade-offs
```

## Design decisions

Full write-up in [`docs/DESIGN.md`](docs/DESIGN.md). The short version:

- **Postgres decides, Redis speeds up.** Correctness never depends on Redis being available; the unique index is the last line of defense.
- **Events instead of direct calls.** Booking returns fast; payment, email and ticket generation scale and fail independently.
- **RabbitMQ over Kafka.** The workload is task distribution, not a replayable event log, so a simple broker fits and is easier to operate.
- **Role checks as a schema directive.** `@hasRole(role: ADMIN)` keeps authorization visible in the schema instead of buried in resolvers.

## Roadmap

- [x] PostgreSQL schema, migrations and seed data
- [x] JWT auth with `@hasRole` directive
- [x] Naive booking baseline and k6 stampede test
- [ ] Concurrency-safe reservation (unique index + Redis hold)
- [ ] `BookingCreated` events and workers (payment, notification, QR)
- [ ] Live seat availability via GraphQL subscriptions
- [ ] DataLoader to remove N+1 queries (before/after query counts)
- [ ] Query depth/complexity limits, Prometheus metrics
- [ ] GitHub Actions CI
- [ ] Public deployment
- [ ] Stretch: LLM tool-calling layer over the API ("find me 2 seats near the front")

## Author

Built by [Vedang](https://github.com/vedang18200).