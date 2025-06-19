// _request.js
/* Centralised fetch helper that prefixes every call with /api */

const API_BASE = '/api'; // ← drop the VITE_GO_API_URL logic here
export async function request(path, opts = {}) {
  console.log('→ [API] ', opts.method || 'GET', path);
  
}

export async function request(path, { method = 'GET', body, headers } = {}) {
  const opts = { method, headers: { ...headers } };

  if (body instanceof FormData) {
    opts.body = body;                         // let the browser set the boundary
  } else if (body !== undefined) {
    opts.body    = JSON.stringify(body);
    opts.headers = { 'Content-Type': 'application/json', ...opts.headers };
  }

  const res  = await fetch(`${API_BASE}${path}`, opts);
  const json = await res.json().catch(() => ({}));

  if (!res.ok) throw new Error(json.message || res.statusText);
  return json;                                // backend already wraps in {data:…}
}
