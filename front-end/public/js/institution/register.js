// public/js/institution/register.js
import { flash } from '../../script.js';
import { registerInstitution } from '../../api/institution.js';

const form = document.querySelector('#register-inst-form');

form.addEventListener('submit', async e => {
  e.preventDefault();

  const name     = form.name.value.trim();
  const director = form.director.value.trim();
  const email    = form.email.value.trim();

  try {
    // pass a plain object—let your _request.js helper JSON.stringify it
    await registerInstitution({ name, email, director });
    flash('Institution registered!');
  } catch (err) {
    flash(err.message);
  }
});
