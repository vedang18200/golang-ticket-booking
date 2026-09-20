import http from 'k6/http';
import { Counter } from 'k6/metrics';

const bookingsOk = new Counter('bookings_ok');
const bookingsConflict = new Counter('bookings_conflict');
const bookingsError = new Counter('bookings_error');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/query';
const EVENT_ID = __ENV.EVENT_ID;
const SEAT_IDS = JSON.parse(__ENV.SEAT_IDS || '[]'); // the 50 seat ids
const USERS = 1000;

export const options = {
  setupTimeout: '5m',
  scenarios: {
    stampede: {
      executor: 'per-vu-iterations',
      vus: USERS,
      iterations: 1,
      maxDuration: '2m',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],
    bookings_error: ['count==0'],
  },
};

const JSON_HEADERS = { 'Content-Type': 'application/json' };

function gql(query, variables, token) {
  const headers = token
    ? Object.assign({ Authorization: `Bearer ${token}` }, JSON_HEADERS)
    : JSON_HEADERS;
  return http.post(BASE_URL, JSON.stringify({ query, variables }), { headers });
}

// Registers all users once, before the stampede starts.
export function setup() {
  const tokens = [];
  for (let i = 0; i < USERS; i++) {
    const res = gql(
      `mutation($e: String!, $p: String!) { register(email: $e, password: $p) { token } }`,
      { e: `load${i}@test.dev`, p: 'password123' },
    );
    tokens.push(res.json('data.register.token'));
  }
  return tokens;
}

export default function (tokens) {
  const seatId = SEAT_IDS[Math.floor(Math.random() * SEAT_IDS.length)];
  const res = gql(
    `mutation($ev: ID!, $s: ID!) { createBooking(eventId: $ev, seatId: $s) { id } }`,
    { ev: EVENT_ID, s: seatId },
    tokens[__VU - 1],
  );

  const body = res.json();
  if (body.data && body.data.createBooking) {
    bookingsOk.add(1);
  } else if (
    body.errors && body.errors[0].extensions &&
    body.errors[0].extensions.code === 'SEAT_TAKEN'
  ) {
    bookingsConflict.add(1);
  } else {
    bookingsError.add(1);
  }
}

// After the run, verify in Postgres: this must return ZERO rows.
//   SELECT seat_id, count(*) FROM bookings
//   WHERE status IN ('PENDING','CONFIRMED') GROUP BY seat_id HAVING count(*) > 1;
