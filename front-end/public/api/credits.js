// credits.js
import { request } from './_request.js';

/**
 * Purchase credits for a given institution.
 * @param {{ name: string, amount: number }} payload
 */
export const purchaseCredits = ({ name, amount }) =>
  request('/purchase', {
    method: 'PATCH',
    body: { name, amount }
  });

export const getMyCredits = () =>
  request('/mycredits');

export const spendCredits = (amount, reason) =>
  request('/spending', {
    method: 'PATCH',
    body: { amount, reason }
  });
