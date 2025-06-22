// front-end/public/js/institution/post-initial.js
//
// Upload INITIAL grades (first declaration)
// ─────────────────────────────────────────────────────────────
import { flash }   from '../../script.js';
import { request } from '../../api/_request.js';   // ← adds the JWT

const form = document.querySelector('#upload-init-form');

form.addEventListener('submit', async (e) => {
  e.preventDefault();

  const fileInput = form.querySelector('input[type="file"]');
  if (!fileInput.files.length) {
    return flash('Please select an XLSX file.');
  }

  const fd = new FormData();
  fd.append('file', fileInput.files[0]);           // name="file" matches Gin handler

  try {
    // The Orchestrator route is a **POST /upload_init** (instructor-only)
    await request('/upload_init', { method: 'POST', body: fd });
    flash('Initial grades uploaded ✔');
  } catch (err) {
    flash(err.message || 'Upload failed');
  }
});
