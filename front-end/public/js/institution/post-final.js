// front-end/public/js/institution/post-final.js
//
// Upload the FINAL grades spreadsheet (credits will be deducted)
// ────────────────────────────────────────────────────────────────
import { flash }    from '../../script.js';
import { request }  from '../../api/_request.js';   // ✅ helper injects JWT

const form = document.querySelector('#upload-final-form');

form.addEventListener('submit', async (e) => {
  e.preventDefault();

  const fileInput = form.querySelector('input[type="file"]');
  if (!fileInput.files.length) {
    return flash('Please select an XLSX file.');
  }

  const fd = new FormData();
  fd.append('file', fileInput.files[0]);            // ↔ name="file" in the form

  try {
    // ────────────────────────────────────────────────────────────
    // •  PATCH /postFinalGrades  (route guarded for instructors)
    // •  request() adds  Authorization: Bearer <jwt>
    // ────────────────────────────────────────────────────────────
    await request('/postFinalGrades', { method: 'PATCH', body: fd });
    flash('Final grades uploaded ✔');
  } catch (err) {
    flash(err.message || 'Upload failed');
  }
});
