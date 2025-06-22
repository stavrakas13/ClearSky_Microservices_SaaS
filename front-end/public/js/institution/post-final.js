// public/js/institution/post-final.js
import { flash } from '../script.js';

const form = document.querySelector('#upload-final-form');

form.addEventListener('submit', async e => {
  e.preventDefault();

  const fileInput = form.querySelector('input[type=file]');
  if (!fileInput.files.length) return flash('Please select an XLSX file.');

  const fd = new FormData();
  fd.append('xlsx', fileInput.files[0]);

  try {
    // Proxy to orchestrator
    const res = await fetch(`${window.API_BASE}/upload_final`, {
      method: 'POST',
      body:   fd,
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.error || body.message || res.statusText);

    flash('Final grades uploaded!');
  } catch (err) {
    flash(err.message);
  }
});
