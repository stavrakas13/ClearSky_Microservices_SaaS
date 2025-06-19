// credits.js
import { request } from './_request.js';

export const purchaseCredits = amount =>
  request('/purchase', { method: 'PATCH', body: { amount } });

export const getMyCredits = () =>
  request('/mycredits');

export const spendCredits = (amount, reason) =>
  request('/spending', { method: 'PATCH', body: { amount, reason } });
