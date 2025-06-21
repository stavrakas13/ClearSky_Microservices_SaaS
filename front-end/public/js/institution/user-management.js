// front-end/public/js/institution/user-management.js
import { flash } from '../../script.js';
import { registerUser } from '../../api/users.js';

const form = document.querySelector('#user-mgmt-form');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const username = form.username.value.trim();
  const password = form.password.value;
  const role     = form.role.value;

  if (!username || !password) {
    return flash('Username and password are required');
  }

  try {
    await registerUser({ username, password, role });
    flash('User added!');
    form.reset();
  } catch (err) {
    flash(`Error: ${err.message}`);
  }
});
