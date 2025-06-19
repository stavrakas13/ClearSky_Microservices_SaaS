// stats.js
import { request } from './_request.js';

export const persistAndCalculateStats = payload =>
  request('/stats/persist', { method: 'POST', body: payload });

export const getDistributions = filters =>
  request('/stats/distributions', { method: 'POST', body: filters });
