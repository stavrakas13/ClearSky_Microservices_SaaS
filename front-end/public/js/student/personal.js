// public/js/student/personal.js
import { flash } from '../script.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = 'spring 2025';
  const user_id     = 'alice';

  try {
    const res = await fetch('/personal/grades', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id, course_id, exam_period }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    document.querySelector('input[readonly][value]').value = body.data.total;
    // populate Q1–Q3 similarly…
  } catch (err) {
    flash(err.message);
  }
});
