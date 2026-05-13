import http from 'k6/http';
import { check, group, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOKEN = __ENV.TOKEN;
const TOKENS = __ENV.TOKENS ? __ENV.TOKENS.split(';').filter((token) => token.length > 0) : [];

export const options = {
  scenarios: {
    steady_gateway_reads: {
      executor: 'ramping-vus',
      stages: [
        { duration: __ENV.RAMP_UP || '15s', target: Number(__ENV.VUS || 25) },
        { duration: __ENV.DURATION || '1m', target: Number(__ENV.VUS || 25) },
        { duration: __ENV.RAMP_DOWN || '15s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<750'],
    checks: ['rate>0.95'],
  },
};

function authHeaders() {
  const token = TOKENS.length > 0 ? TOKENS[(__VU - 1) % TOKENS.length] : TOKEN;
  return {
    Authorization: `Bearer ${token}`,
    'X-Request-ID': `load-read-${__VU}-${__ITER}`,
    'X-Correlation-ID': 'load-read',
  };
}

export default function () {
  if (!TOKEN && TOKENS.length === 0) {
    throw new Error('TOKEN or TOKENS env var is required');
  }

  group('gateway protected reads', () => {
    const profile = http.get(`${BASE_URL}/api/v1/users/me`, { headers: authHeaders() });
    check(profile, {
      'profile status is 200': (r) => r.status === 200,
      'profile success envelope': (r) => r.json('success') === true,
    });

    const feed = http.get(`${BASE_URL}/api/v1/feed`, { headers: authHeaders() });
    check(feed, {
      'feed status is 200': (r) => r.status === 200,
      'feed success envelope': (r) => r.json('success') === true,
    });

    const notifications = http.get(`${BASE_URL}/api/v1/notifications`, { headers: authHeaders() });
    check(notifications, {
      'notifications status is 200': (r) => r.status === 200,
      'notifications success envelope': (r) => r.json('success') === true,
    });
  });

  sleep(Number(__ENV.SLEEP_SECONDS || 1));
}
