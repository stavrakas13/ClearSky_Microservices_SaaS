// student/review-request.js
import { flash } from '../../script.js';
import { postReviewRequest } from '../../api/student.js';

const form = document.querySelector('form[action="/api/appeals"]');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const params       = new URLSearchParams(location.search);
  const course_id    = params.get('course');
  const exam_period  = 'spring 2025';
  const user_id      = 'alice';
  const student_message = form.message.value.trim();

  try {
    await postReviewRequest({ user_id, course_id, exam_period, student_message });
    flash('Review request submitted!');
  } catch (err) {
    flash(err.message);
  }
});
