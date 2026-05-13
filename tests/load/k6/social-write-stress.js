import http from 'k6/http';
import { check, group, sleep } from 'k6';

http.setResponseCallback(http.expectedStatuses({ min: 200, max: 399 }, 403));

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const ACTOR_TOKEN = __ENV.ACTOR_TOKEN;
const ACTOR_TOKENS = __ENV.ACTOR_TOKENS ? __ENV.ACTOR_TOKENS.split(';').filter((token) => token.length > 0) : [];
const AUTHOR_TOKEN = __ENV.AUTHOR_TOKEN;
const POST_ID = __ENV.POST_ID;

export const options = {
  scenarios: {
    post_like_stress: {
      executor: 'ramping-vus',
      stages: [
        { duration: __ENV.RAMP_UP || '10s', target: Number(__ENV.VUS || 10) },
        { duration: __ENV.DURATION || '45s', target: Number(__ENV.VUS || 10) },
        { duration: __ENV.RAMP_DOWN || '10s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.10'],
    http_req_duration: ['p(95)<1000'],
    checks: ['rate>0.90'],
  },
};

function headers(token, correlationID) {
  return {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
    'X-Request-ID': `load-write-${__VU}-${__ITER}`,
    'X-Correlation-ID': correlationID,
  };
}

export default function () {
  if (!ACTOR_TOKEN || !AUTHOR_TOKEN || !POST_ID) {
    if (ACTOR_TOKENS.length === 0 || !AUTHOR_TOKEN || !POST_ID) {
      throw new Error('ACTOR_TOKEN or ACTOR_TOKENS, AUTHOR_TOKEN, and POST_ID env vars are required');
    }
  }
  const actorToken = ACTOR_TOKENS.length > 0 ? ACTOR_TOKENS[(__VU - 1) % ACTOR_TOKENS.length] : ACTOR_TOKEN;

  group('post interaction write path', () => {
    const interaction = http.post(
      `${BASE_URL}/api/v1/posts/${POST_ID}/interactions`,
      JSON.stringify({ interaction_type: 'like' }),
      { headers: headers(actorToken, 'load-write-like') },
    );
    check(interaction, {
      'interaction accepted or forbidden duplicate': (r) => r.status === 202 || r.status === 403,
      'interaction response is bounded': (r) => r.timings.duration < 1000,
    });

    if (__VU === 1) {
      const notifications = http.get(`${BASE_URL}/api/v1/notifications`, {
        headers: headers(AUTHOR_TOKEN, 'load-write-notifications'),
      });
      check(notifications, {
        'author notifications status is 200': (r) => r.status === 200,
        'author notifications envelope': (r) => r.json('success') === true,
      });
    }
  });

  sleep(Number(__ENV.SLEEP_SECONDS || 1));
}
