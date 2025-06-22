import { flash } from '../../script.js';
import { registerUser, changePassword } from '../../api/users.js';

const form            = document.querySelector('#user-mgmt-form');
const roleSelect      = document.querySelector('#role');
const studentIdGroup  = document.querySelector('#student-id-group');

// Show/hide Student ID field
roleSelect.addEventListener('change', () => {
  studentIdGroup.style.display =
    roleSelect.value === 'student' ? 'block' : 'none';
});

// Add user handler
form.addEventListener('submit', async e => {
  e.preventDefault();

  const username   = form.username.value.trim();
  const password   = form.password.value;
  const role       = form.role.value;
  const student_id = role === 'student'
    ? form.student_id.value.trim()
    : undefined;

  if (!username || !password) {
    return flash('Username and password are required');
  }

  try {
    await registerUser({ username, password, role, student_id });
    flash('User added!');
    form.reset();
    studentIdGroup.style.display = 'none';
  } catch (err) {
    flash(`Error: ${err.message}`);
  }
});

// Change password handler
const changeForm = document.querySelector('#change-pass-form');
changeForm.addEventListener('submit', async e => {
  e.preventDefault();

  const payload = {
    username     : changeForm.username.value.trim(),
    old_password : changeForm.old_password.value,
    new_password : changeForm.new_password.value
  };

  try {
    await changePassword(payload);
    flash('Password changed ✔');
    changeForm.reset();
  } catch (err) {
    flash(`Error: ${err.message}`);
  }
});
