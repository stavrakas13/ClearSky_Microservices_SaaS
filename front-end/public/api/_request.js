// front-end/public/api/_request.js

// NOTE: in your browser, "orchestrator" isn't a DNS name.
// Use localhost:8080 (or adjust if you run the orchestrator elsewhere).
const API_BASE = 'http://localhost:8080';

/**
 * Read the JWT from localStorage (or cookie fallback).
 */
function getJWT() {
  const fromLS = window.localStorage?.getItem('jwt');
  if (fromLS) return fromLS;
  const m = document.cookie.match(/(?:^|;\s*)jwt=([^;]+)/);
  return m ? decodeURIComponent(m[1]) : null;
}

/**
 * Generic request helper that:
 *  • automatically JSON‐stringifies objects (except GETs, which become query strings)
 *  • sends FormData unchanged
 *  • injects `Authorization: Bearer <token>` if you have a JWT
 */
export async function request(path, { method = 'GET', body, headers } = {}) {
  console.log('→ [API]', method, path, 'body:', body);

  // build full URL
  let url = API_BASE + path;
  const opts = { method, headers: { ...headers } };

  // inject auth
  const token = getJWT();
  if (token && !opts.headers.Authorization) {
    opts.headers.Authorization = `Bearer ${token}`;
  }

  // If GET + plain object, turn into query string
  if (method.toUpperCase() === 'GET' && body && !(body instanceof FormData)) {
    const qs = new URLSearchParams(body).toString();
    url += (url.includes('?') ? '&' : '?') + qs;
  }
  // If it's FormData, send as-is
  else if (body instanceof FormData) {
    opts.body = body;
  }
  // Otherwise (non-GET JSON payload)
  else if (body !== undefined) {
    opts.body = JSON.stringify(body);
    opts.headers = { 'Content-Type': 'application/json', ...opts.headers };
  }

  const res = await fetch(url, opts);
  const json = await res.json().catch(() => ({}));

  if (!res.ok) {
    throw new Error(json.message || json.error || res.statusText);
  }
  return json;
}
