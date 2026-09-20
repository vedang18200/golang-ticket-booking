CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role          text NOT NULL DEFAULT 'USER' CHECK (role IN ('USER', 'ADMIN')),
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE venues (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    city text NOT NULL
);

CREATE TABLE seats (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id    uuid NOT NULL REFERENCES venues(id),
    section     text NOT NULL,
    row_label   text NOT NULL,
    seat_number int  NOT NULL,
    UNIQUE (venue_id, section, row_label, seat_number)
);

CREATE TABLE events (
    id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id  uuid NOT NULL REFERENCES venues(id),
    title     text NOT NULL,
    starts_at timestamptz NOT NULL
);

CREATE TABLE bookings (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id),
    event_id   uuid NOT NULL REFERENCES events(id),
    seat_id    uuid NOT NULL REFERENCES seats(id),
    status     text NOT NULL DEFAULT 'PENDING'
               CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'EXPIRED')),
    expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX bookings_user_idx ON bookings (user_id);
