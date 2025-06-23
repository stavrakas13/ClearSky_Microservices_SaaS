// public/api/stats.js
import { request } from './_request.js';

/**
 * GET /stats/available
 * returns either { data: […] } or […] directly
 */
export async function getAvailableStats() {
  const res = await request('/stats/available');
  // if your API wraps in { data: […] }, use that, else assume res itself is the array
  return res.data ?? res;
}

/**
 * POST /stats/distributions
 * again, unwrap .data if present
 */
export async function getDistributions({ course, declarationPeriod, classTitle }) {
  const res = await request('/stats/distributions', {
    method: 'POST',
    body: { course, declarationPeriod, classTitle }
  });
  return res.data ?? res;
}
