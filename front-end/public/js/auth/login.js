// public/js/auth/login.js
import { flash } from '../../script.js';
import { loginUser } from '../../api/users.js';

const form     = document.querySelector('#login-form');
const errorMsg = document.querySelector('#error-msg');

form.addEventListener('submit', async e => {
  e.preventDefault();
  errorMsg.style.display = 'none';

  const input    = form.username.value.trim();
  const password = form.password.value;
  const isEmail  = /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(input);
  const payload  = isEmail
    ? { email: input, password }
    : { username: input, password };

  try {
    // loginUser now returns { role, token }
    const { role, token } = await loginUser(payload);

    // store the JWT so our API helpers will use it
    localStorage.setItem('jwt', token);

    // Redirect based on role
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
