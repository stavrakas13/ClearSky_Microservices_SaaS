// public/js/institution/post-initial.js
import { flash } from '../../script.js';

const form = document.querySelector('#upload-init-form');

form.addEventListener('submit', async e => {
  e.preventDefault();

  const fileInput = form.querySelector('input[type=file]');
  if (!fileInput.files.length) {
    return flash('Please select an XLSX file.');
  }

  const fd = new FormData();
  fd.append('file', fileInput.files[0]);  // matches name="file" on the form

  try {
    // Now call the orchestrator directly:
    const res = await fetch(`${window.API_BASE}/upload_init`, {
      method: 'POST',
      body:   fd,
    });

    const body = await res.json();
    if (!res.ok) throw new Error(body.error || body.message || res.statusText);

    flash('Initial grades uploaded!');
  } catch (err) {
    flash(err.message);
  }
});
