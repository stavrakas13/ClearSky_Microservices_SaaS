// front-end/public/api/_request.js
const API_BASE = '/api';

export async function request(path, { method = 'GET', body, headers } = {}) {
  // Debug log every API call
  console.log('→ [API]', method, path, 'body:', body);

  // Pull JWT from localStorage (if any)
  const token = localStorage.getItem('jwt');

  // Build options, injecting Authorization header if we have a token
  const opts = {
    method,
    headers: {
      ...headers,
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
  };

  if (body instanceof FormData) {
    // let the browser set the multipart boundary
    opts.body = body;
  } else if (body !== undefined) {
    opts.body = JSON.stringify(body);
    opts.headers = {
      'Content-Type': 'application/json',
      ...opts.headers
    };
  }

  // Send credentials (cookies) too, if you ever need them
  const res = await fetch(`${API_BASE}${path}`, {
    ...opts,
    credentials: 'include'
  });

  const json = await res.json().catch(() => ({}));

  if (!res.ok) throw new Error(json.message || res.statusText);
  return json; // backend wraps responses in { data: … }
}
