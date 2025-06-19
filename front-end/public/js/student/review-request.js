// public/js/student/review-request.js
import { flash } from '../script.js';

const form = document.querySelector('form[action="/api/appeals"]');
form.addEventListener('submit', async e => {
  e.preventDefault();
  const params = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = 'spring 2025';
  const user_id     = 'alice';
  const student_message = form.message.value.trim();

  try {
    const res = await fetch('/student/reviewRequest', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ course_id, user_id, exam_period, student_message }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    flash('Review request submitted!');
  } catch (err) {
    flash(err.message);
  }
});
