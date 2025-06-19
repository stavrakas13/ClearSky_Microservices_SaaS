// public/js/student/review-status.js
import { flash } from '../script.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = 'spring 2025';
  const user_id     = 'alice';

  try {
    const res = await fetch('/student/status', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ course_id, user_id, exam_period }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    document.querySelector('textarea[readonly]').textContent =
      body.data.instructor_message;
  } catch (err) {
    flash(err.message);
  }
});
