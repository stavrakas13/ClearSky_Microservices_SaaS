// institution/register.js
import { flash } from '../../script.js';
import { registerInstitution } from '../../api/institution.js';

const form = document.querySelector('#register-inst-form');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const name   = form.name.value.trim();
  const domain = form.domain.value.trim();
  const email  = form.email.value.trim();

  try {
    await registerInstitution({ name, domain, email });
    flash('Institution registered!');
  } catch (err) {
    flash(err.message);
  }
});
