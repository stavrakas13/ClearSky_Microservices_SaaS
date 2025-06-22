// front-end/public/js/student/review-status.js
import { flash } from '../../script.js';
import { getReviewStatus } from '../../api/student.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params      = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = params.get('period');   // optional

  if (!course_id) {
    flash('Please pick a course first');
    window.location.href = '/student/my-courses';
    return;
  }

  try {
    const { data } = await getReviewStatus({ course_id, exam_period });
    document.querySelector('textarea[readonly]').textContent =
      data.instructor_message || 'No response yet.';
  } catch (err) {
    flash(err.message);
  }
});
