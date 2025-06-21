import { flash } from '../../script.js';
import { loginUser } from '../../api/users.js';

const form     = document.querySelector('#login-form');
const errorMsg = document.querySelector('#error-msg');

form.addEventListener('submit', async e => {
  e.preventDefault();
  errorMsg.style.display = 'none';

  const username = form.username.value.trim();
  const password = form.password.value;

  try {
    await loginUser({ username, password });
    // JWT is in localStorage, redirect now
    window.location.href = '/';
  } catch (err) {
    errorMsg.textContent        = err.message;
    errorMsg.style.display      = 'block';
  }
});
