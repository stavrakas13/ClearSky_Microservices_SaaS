// public/api.js
const API_BASE = ''; // if your Go server is on the same host/port; otherwise set to e.g. 'http://localhost:3000'

async function handleResponse(res) {
  if (!res.ok) {
    let err = { message: res.statusText };
    try { err = await res.json(); } catch {}
    throw new Error(err.message || res.statusText);
  }
  return res.json();
}

// === Credits ===
export function purchaseCredits(amount) {
  return fetch(`${API_BASE}/purchase`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ amount }),
  }).then(handleResponse);
}

export function getMyCredits() {
  return fetch(`${API_BASE}/mycredits`, {
    method: 'GET',
  }).then(handleResponse);
}

export function spendCredits(amount, reason) {
  return fetch(`${API_BASE}/spending`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ amount, reason }),
  }).then(handleResponse);
}

// === Institution ===
export function registerInstitution({ name, domain, email }) {
  return fetch(`${API_BASE}/registration`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, domain, email }),
  }).then(handleResponse);
}

// === Excel upload ===
export function uploadInitialGrades(file) {
  const fd = new FormData();
  fd.append('xlsx', file);
  return fetch(`${API_BASE}/upload_init`, {
    method: 'POST',
    body: fd,
  }).then(handleResponse);
}

// === Statistics ===
export function persistAndCalculateStats(payload) {
  return fetch(`${API_BASE}/stats/persist`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  }).then(handleResponse);
}

export function getDistributions(payload) {
  return fetch(`${API_BASE}/stats/distributions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  }).then(handleResponse);
}

// === Personal (student) ===
export function getStudentCourses({ user_id }) {
  return fetch(`${API_BASE}/personal/courses`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id }),
  }).then(handleResponse);
}

export function getPersonalGrades({ user_id, course_id, exam_period }) {
  return fetch(`${API_BASE}/personal/grades`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id, course_id, exam_period }),
  }).then(handleResponse);
}

// === Student review ===
export function postReviewRequest({ user_id, course_id, exam_period, student_message }) {
  return fetch(`${API_BASE}/student/reviewRequest`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id, course_id, exam_period, student_message }),
  }).then(handleResponse);
}

export function getReviewStatus({ user_id, course_id, exam_period }) {
  return fetch(`${API_BASE}/student/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id, course_id, exam_period }),
  }).then(handleResponse);
}

// === Instructor review ===
export function getPendingReviews({ course_id, exam_period }) {
  return fetch(`${API_BASE}/instructor/review-list`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ course_id, exam_period }),
  }).then(handleResponse);
}

export function postInstructorReply({ user_id, course_id, exam_period, instructor_reply_message, instructor_action }) {
  return fetch(`${API_BASE}/instructor/reply`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id, course_id, exam_period, instructor_reply_message, instructor_action }),
  }).then(handleResponse);
}

// === User management ===
export function registerUser(payload) {
  return fetch(`${API_BASE}/user/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  }).then(handleResponse);
}

export function loginUser({ username, password }) {
  return fetch(`${API_BASE}/user/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  }).then(handleResponse);
}

export function deleteUser({ user_id }) {
  return fetch(`${API_BASE}/user/delete`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id }),
  }).then(handleResponse);
}

export function googleLoginUser(token) {
  return fetch(`${API_BASE}/user/google-login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  }).then(handleResponse);
}
