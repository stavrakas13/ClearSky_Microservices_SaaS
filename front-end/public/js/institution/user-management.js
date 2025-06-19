// institution/user-management.js
import { flash } from '../../script.js';
import { registerUser } from '../../api/users.js';

const form = document.querySelector('#user-mgmt-form');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const role = form.role.value;
  const id   = form.id.value.trim();

  try {
    await registerUser({ role, id });
    flash('User added!');
  } catch (err) {
    flash(err.message);
  }
});
