// file: front-end/public/api/_request.js

// Base URL of your Go API
const API_BASE = window.GO_API_URL || 'http://localhost:8080';

export async function request(path, { method = 'GET', body, headers } = {}) {
  console.log('→ [API]', method, path, 'body:', body);

  const opts = { method, headers: { ...headers } };

  if (body instanceof FormData) {
    // Let the browser set the multipart boundary for form data
    opts.body = body;
  } else if (body !== undefined) {
    opts.body    = JSON.stringify(body);
    opts.headers = { 'Content-Type': 'application/json', ...opts.headers };
  }

  const res = await fetch(`${API_BASE}${path}`, opts);
  const json = await res.json().catch(() => ({}));

  if (!res.ok) {
    throw new Error(json.message || res.statusText);
  }

  return json; // Expecting { data: … } shape from your API
}
