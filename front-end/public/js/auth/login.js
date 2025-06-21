// file: front-end/public/js/auth/login.js
import { flash } from '../../script.js';
import { loginUser, googleLoginUser } from '../../api/users.js';

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
    // loginUser now returns the full response { role, status, token, userId }
    const { role } = await loginUser(payload);

    // Redirect based on role:
    if (role === 'institution_representative') {
      window.location.href = '/institution';
    } else if (role === 'instructor') {
      window.location.href = '/instructor';
    } else if (role === 'student') {
      window.location.href = '/student';
    } else {
      // fallback
      window.location.href = '/';
    }
  } catch (err) {
    errorMsg.textContent = err.message;
    errorMsg.style.display = 'block';
  }
});

// Optional: Google login
const googleBtn = document.querySelector('#google-login-btn');
if (googleBtn) {
  googleBtn.addEventListener('click', async () => {
    try {
      const googleToken = await getGoogleOAuthTokenSomehow();
      const { role }   = await googleLoginUser(googleToken);
      if (role === 'institution_representative') window.location.href = '/institution';
      else if (role === 'instructor')             window.location.href = '/instructor';
      else if (role === 'student')                window.location.href = '/student';
      else                                        window.location.href = '/';
    } catch (err) {
      flash(`Google login failed: ${err.message}`);
    }
  });
}
