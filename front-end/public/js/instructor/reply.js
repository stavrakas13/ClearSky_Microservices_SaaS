// instructor/reply.js
import { flash } from '../../script.js';
import { postInstructorReply } from '../../api/instructor.js';

const form = document.querySelector('form[action="/api/appeals/reply"]');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const params = new URLSearchParams(location.search);
  const reqId = params.get('req');

  const course_id   = 'software II';
  const exam_period = 'spring 2025';
  const user_id     = reqId;

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
  } catch (err) {
    flash(err.message);
  }
});
