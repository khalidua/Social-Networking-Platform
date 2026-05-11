import http from 'k6/http';
import { check, sleep } from 'k6';


const TOKEN = __ENV.TOKEN;
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const FEED_URL = `${BASE_URL}/api/v1/feed`;

export const options = {
  vus: 50,
  duration: '30s',
};

export default function () {
  const res = http.get(FEED_URL, {
    headers: {
      'Authorization': `Bearer ${TOKEN}`,
      'X-User-ID': 'user123',
    },
  });

  if (__VU === 1 && __ITER === 0 && res.status !== 200) {
    console.error(`Non-200 response: status=${res.status}, body=${res.body}`);
  }

  check(res, {
    'status is 200': (r) => r.status === 200,
    'latency under 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}