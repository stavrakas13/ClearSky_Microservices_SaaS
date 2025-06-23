// front-end/public/js/instructor/reply.js
import { flash } from '../../script.js';
import { postInstructorReply, getPendingReviews } from '../../api/instructor.js';

const form = document.querySelector('#instructor-reply-form');

window.addEventListener('DOMContentLoaded', async () => {
  const qs = new URLSearchParams(location.search);
  const user_id     = qs.get('student');
  const course_id   = qs.get('course');
  const exam_period = qs.get('period'); // may be null

  if (!user_id || !course_id) {
    flash('Missing URL parameters – please navigate from the review list.');
    return;
  }

  try {
    // Fetch the pending review for this student/course/period
    const reviews = await getPendingReviews({
      course_id,
      exam_period,
    });
    // Find the review for this student
    const review = reviews.find(
      (r) =>
        r.student_id === user_id &&
        r.course_id === course_id &&
        (exam_period ? r.exam_period === exam_period : true)
    );
    if (!review) {
      flash('No review request found for this student.');
      return;
    }
    // Fill in the details at the top of the form
    document.getElementById('course_id').value = review.course_id ?? '';
    document.getElementById('exam_period').value = review.exam_period ?? '';
    document.getElementById('student_id').value = review.student_id ?? '';
    document.getElementById('student_message').value = review.student_message ?? '';
    document.getElementById('review_created_at').value = review.review_created_at
      ? new Date(review.review_created_at).toLocaleString()
      : '';
  } catch (err) {
    flash(err.message || 'Failed to fetch review details');
  }
});

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
