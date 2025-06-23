// front-end/public/js/instructor/reply.js
import { flash } from '../../script.js';
import { postInstructorReply } from '../../api/instructor.js';

const form = document.querySelector('#instructor-reply-form');

form.addEventListener('submit', async (e) => {
  e.preventDefault();

  const qs = new URLSearchParams(location.search);
  const user_id     = qs.get('student');
  const course_id   = qs.get('course');
  const exam_period = qs.get('period');          // may be null
  const instructor_reply_message = form.message.value.trim();
  const instructor_action        = form.decision.value;

  if (!user_id || !course_id) {
    flash('Missing URL parameters – please navigate from the review list.');
    return;
  }

  try {
    await postInstructorReply({
      user_id,
      course_id,
      exam_period,
      instructor_reply_message,
      instructor_action
    });
    flash('Reply sent!');
    /* Redirect back to the list after a short toast. */
    setTimeout(() => (window.location.href = '/instructor/review-list'), 1200);
  } catch (err) {
    flash(err.message || 'Failed to send reply');
  }
});
