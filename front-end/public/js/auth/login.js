// front-end/public/js/auth/login.js
import { flash } from '../../script.js';
import { loginUser } from '../../api/users.js';

const form     = document.querySelector('#login-form');
const errorMsg = document.querySelector('#error-msg');

form.addEventListener('submit', async e => {
  e.preventDefault();
  errorMsg.style.display = 'none';

  // ──────────────────────────────────────────────────────────────
  // 1) Build payload (username or e-mail)
  // ──────────────────────────────────────────────────────────────
  const input    = form.username.value.trim();
  const password = form.password.value;
  const isEmail  = /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(input);
  const payload  = isEmail
    ? { email: input, password }
    : { username: input, password };

  try {
    // ────────────────────────────────────────────────────────────
    // 2) Ask orchestrator to log us in → { role, token }
    // ────────────────────────────────────────────────────────────
    const { role, token } = await loginUser(payload);

    // 3) Persist JWT so every future fetch() carries Authorization: Bearer …
    localStorage.setItem('jwt', token);

    // 4) Tell the Express layer to remember who we are (for EJS templates)
    await fetch('/api/session', {
      method : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body   : JSON.stringify({
        username: input,
        role
      })
    });

    // ────────────────────────────────────────────────────────────
    // 5) Redirect according to role
    // ────────────────────────────────────────────────────────────
    if (['institution_representative', 'representative'].includes(role)) {
      window.location.href = '/institution';
    } else if (role === 'instructor') {
      window.location.href = '/instructor';
    } else if (role === 'student') {
      window.location.href = '/student';
    } else {
      window.location.href = '/';
    }
  } catch (err) {
    errorMsg.textContent = err.message;
    errorMsg.style.display = 'block';
  }
});
