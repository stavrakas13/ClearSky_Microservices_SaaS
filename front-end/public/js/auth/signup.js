// auth/signup.js
import { flash } from '../../script.js';
import { registerUser } from '../../api/users.js';

const form = document.querySelector('main form');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const role     = form.role.value;
  const username = form.username.value.trim();
  const password = form.password.value;

  if (!username || !password) {
    return flash('Username and password are required');
  }
  try {
    await registerUser({ username, password, role });
    flash('Signup successful! Redirecting to login…');
    setTimeout(() => (window.location.href = '/login'), 1500);
  } catch (err) {
    flash(`Error: ${err.message}`);
  }
});
