// public/js/institution/register.js
import { flash } from '../script.js';

const form = document.querySelector('fieldset form');
form.addEventListener('submit', async e => {
  e.preventDefault();
  const name   = form.name.value.trim();
  const domain = form.domain.value.trim();
  const email  = form.email.value.trim();

  try {
    const res = await fetch('/registration', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, domain, email }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    flash('Institution registered!');
  } catch (err) {
    flash(err.message);
  }
});
