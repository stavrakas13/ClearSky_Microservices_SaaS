// public/js/auth/login.js
import { flash } from '../script.js';

const form = document.querySelector('main form');
form.addEventListener('submit', async e => {
  e.preventDefault();
  const username = form.username.value.trim();
  const password = form.password.value;

  try {
    const res = await fetch('/user/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    window.location.href = '/';
  } catch (err) {
    flash(err.message);
  }
});
