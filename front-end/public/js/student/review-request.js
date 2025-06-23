// front-end/public/js/student/review-request.js
import { flash } from '../../script.js';
import { postReviewRequest } from '../../api/student.js';
  
const form = document.querySelector('#student-review-request-form');
if (!form) {
  console.error('Student review request form (#student-review-request-form) not found!');
} else {
  form.addEventListener('submit', async e => {
    e.preventDefault();

    const params          = new URLSearchParams(location.search);
    const course_id       = params.get('course');
    const exam_period     = params.get('period') || undefined;   // allow backend to choose
    const student_message = form.message.value.trim();

    if (!course_id) {
      flash('Please pick a course first');
      window.location.href = '/student/my-courses';
      return;
    }

    try {
      await postReviewRequest({ course_id, exam_period, student_message });
      flash('Review request submitted!');
    } catch (err) {
      flash(err.message);
    }
  });
}
