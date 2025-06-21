// front-end/public/js/auth/logout.js
import { flash } from '../../script.js';

const logoutBtn = document.querySelector('#logout-button');
if (logoutBtn) {
  logoutBtn.addEventListener('click', e => {
    e.preventDefault();
    // Remove JWT so future API calls are unauthenticated
    localStorage.removeItem('jwt');
    flash('Logged out');
    window.location.href = '/login';
  });
}
