// public/js/instructor/post-initial.js
import { flash } from '../script.js';

const form = document.querySelector('form[action="/api/grades/upload"]') ||
             document.querySelector('form[action="/api/grades/upload"]');
form.addEventListener('submit', async e => {
  e.preventDefault();
  const fileInput = form.querySelector('input[type=file]');
  if (!fileInput.files.length) return flash('Please select an XLSX file.');

  const fd = new FormData();
  fd.append('xlsx', fileInput.files[0]);

  try {
    const res = await fetch('/upload_init', {
      method: 'POST',
      body: fd,
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    flash('Initial grades uploaded!');
  } catch (err) {
    flash(err.message);
  }
});
