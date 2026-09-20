CREATE UNIQUE INDEX bookings_one_active_per_seat
    ON bookings (event_id, seat_id)
    WHERE status IN ('PENDING', 'CONFIRMED');
