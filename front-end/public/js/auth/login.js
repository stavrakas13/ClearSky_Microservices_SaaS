// auth/login.js
import { flash } from '../../script.js';
import { loginUser } from '../../api/users.js';

const form = document.querySelector('main form');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const username = form.username.value.trim();
  const password = form.password.value;

  try {
    await loginUser({ username, password });
    window.location.href = '/';
  } catch (err) {
    flash(err.message);
  }
});
