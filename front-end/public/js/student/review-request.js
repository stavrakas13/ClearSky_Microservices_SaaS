// front-end/public/js/student/review-request.js
//
// Submits a grade-review request. Ensures `course_id` contains
// only the numeric part before sending it to /student/reviewRequest.
// ────────────────────────────────────────────────────────────────────
import { flash } from '../../script.js';
import { postReviewRequest } from '../../api/student.js';

const form = document.querySelector('#student-review-request-form');
if (!form) {
  console.error('Student review request form (#student-review-request-form) not found!');
} else {
  form.addEventListener('submit', async e => {
    e.preventDefault();

    const params = new URLSearchParams(location.search);

    /* ── 1) Extract & sanitise course_id ─────────────────────────── */
    let course_id = params.get('course');                 // could be “ΤΕΧΝ… (3205)”
    if (course_id) {
      course_id = course_id.match(/\((\d+)\)/)?.[1]       // digits inside (...)
               || course_id.replace(/\D/g, '');           // any digits
    }

    const exam_period     = params.get('period') || undefined;
    const student_message = form.message.value.trim();

    if (!course_id) {
      flash('Course id missing – please navigate from “My courses” page.');
      location.href = '/student/my-courses';
      return;
    }

    try {
      await postReviewRequest({ course_id, exam_period, student_message });
      flash('Review request submitted!');
      form.reset();
    } catch (err) {
      flash(err.message || 'Failed to submit review request');
    }
  });
}
