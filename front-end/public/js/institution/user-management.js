// public/js/institution/user-management.js
import { flash } from '../script.js';

const form = document.querySelector('fieldset form');
form.addEventListener('submit', async e => {
  e.preventDefault();
  const role = form.role.value;
  const id   = form.id.value.trim();

  try {
    const res = await fetch('/user/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ role, id }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    flash('User added!');
  } catch (err) {
    flash(err.message);
  }
});
