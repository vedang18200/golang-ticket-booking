INSERT INTO venues (name, city) VALUES ('Test Arena', 'Mumbai');

INSERT INTO seats (venue_id, section, row_label, seat_number)
SELECT v.id, 'A', 'R' || ((n - 1) / 10 + 1), ((n - 1) % 10) + 1
FROM venues v, generate_series(1, 50) AS n
WHERE v.name = 'Test Arena';

INSERT INTO events (venue_id, title, starts_at)
SELECT id, 'Load Test Show', now() + interval '7 days'
FROM venues WHERE name = 'Test Arena';
