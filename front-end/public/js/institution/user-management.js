import { flash } from '../../script.js';
import { registerUser } from '../../api/users.js';

const form            = document.querySelector('#user-mgmt-form');
const roleSelect      = document.querySelector('#role');
const studentIdGroup  = document.querySelector('#student-id-group');

roleSelect.addEventListener('change', () => {
  studentIdGroup.style.display =
    roleSelect.value === 'student' ? 'block' : 'none';
});

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
