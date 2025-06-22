// public/js/instructor/review-list.js
import { flash } from '../script.js';
import { getPendingReviews } from '../../api/instructor.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const payload = { course_id: 'software II', exam_period: 'spring 2025' };
    // API now returns the array directly
    const reviews = await getPendingReviews(payload);

    const tbody = document.querySelector('table tbody');
    tbody.innerHTML = reviews
      .map(r => `
        <tr>
          <td>${r.course_name}</td>
          <td>${r.exam_period}</td>
          <td>${r.student}</td>
          <td>
            <a class="button" href="/instructor/reply?req=${r.id}">Reply</a>
          </td>
        </tr>
      `)
      .join('');
  } catch (err) {
    flash(err.message);
  }
});
