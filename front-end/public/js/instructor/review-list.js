// front-end/public/js/instructor/review-list.js
import { flash } from '../../script.js';
import { getPendingReviews } from '../../api/instructor.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    /* Supply filters here if you have a selector UI; blank object = “all”. */
    const reviews = await getPendingReviews({});

    const tbody = document.querySelector('table tbody');
    if (!reviews.length) {
      tbody.innerHTML =
        '<tr><td colspan="4" style="text-align:center;">No pending requests 💤</td></tr>';
      return;
    }

    /* Build table rows. */
    tbody.innerHTML = reviews
      .map((r) => {
        const qs = new URLSearchParams({
          student: r.student_id,
          course: r.course_id,
          period: r.exam_period ?? ''
        }).toString();

        return `
          <tr>
            <td>${r.course_id}</td>
            <td>${r.exam_period ?? '-'}</td>
            <td>${r.student_id}</td>
            <td>
              <a class="button" href="/instructor/reply?${qs}">Reply</a>
            </td>
          </tr>
        `;
      })
      .join('');
  } catch (err) {
    flash(err.message || 'Failed to fetch review list');
  }
});
