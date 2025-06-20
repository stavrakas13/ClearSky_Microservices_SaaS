const API_BASE = '/api';

export async function request(path, { method = 'GET', body, headers } = {}) {
  // Debug log every API call
  console.log('→ [API]', method, path, 'body:', body);

  const opts = { method, headers: { ...headers } };

  if (body instanceof FormData) {
    // let the browser set the multipart boundary
    opts.body = body;
  } else if (body !== undefined) {
    opts.body    = JSON.stringify(body);
    opts.headers = { 'Content-Type': 'application/json', ...opts.headers };
  }

  const res  = await fetch(`${API_BASE}${path}`, opts);
  const json = await res.json().catch(() => ({}));

  if (!res.ok) throw new Error(json.message || res.statusText);
  return json; // backend wraps responses in { data: … }
}
