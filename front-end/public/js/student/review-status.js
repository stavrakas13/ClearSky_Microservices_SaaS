// student/review-status.js
import { flash } from '../../script.js';
import { getReviewStatus } from '../../api/student.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params      = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = 'spring 2025'; // or dynamic

  try {
    const { data } = await getReviewStatus({ course_id, exam_period });
    document.querySelector('textarea[readonly]').textContent = data.instructor_message || 'No response yet.';
  } catch (err) {
    flash(err.message);
  }
});
