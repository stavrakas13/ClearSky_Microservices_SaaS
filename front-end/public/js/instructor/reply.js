// public/js/instructor/reply.js
import { flash } from '../../script.js';
import { postInstructorReply } from '../../api/instructor.js';

const form = document.querySelector('#instructor-reply-form');

form.addEventListener('submit', async e => {
  e.preventDefault();

  const params       = new URLSearchParams(location.search);
  const user_id      = params.get('req');                // the student’s ID
  const course_id    = 'software II';                    // or pull from hidden input
  const exam_period  = 'spring 2025';                    // likewise
  const instructor_reply_message = form.message.value.trim();
  const instructor_action        = form.decision.value;

  try {
    await postInstructorReply({
      user_id,
      course_id,
      exam_period,
      instructor_reply_message,
      instructor_action
    });
    flash('Reply sent!');
    // Optionally redirect back to the list:
    // window.location.href = '/instructor/review-list';
  } catch (err) {
    flash(err.message);
  }
});
