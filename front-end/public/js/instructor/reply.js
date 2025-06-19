// public/js/instructor/reply.js
import { flash } from '../script.js';

const form = document.querySelector('form[action="/api/appeals/reply"]');
form.addEventListener('submit', async e => {
  e.preventDefault();
  const params = new URLSearchParams(location.search);
  const reqId = params.get('req');

  // derive these from your page or hidden fields
  const course_id  = 'software II';
  const exam_period= 'spring 2025';
  const user_id    = reqId; // or extract real user_id

  const instructor_reply_message = form.message.value.trim();
  const instructor_action        = form.decision.value;

  try {
    const res = await fetch('/instructor/reply', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        course_id,
        user_id,
        exam_period,
        instructor_reply_message,
        instructor_action,
      }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    flash('Reply sent!');
  } catch (err) {
    flash(err.message);
  }
});
