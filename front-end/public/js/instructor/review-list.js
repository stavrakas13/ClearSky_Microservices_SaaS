// instructor/review-list.js
import { flash } from '../../script.js';
import { getPendingReviews } from '../../api/instructor.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const { data } = await getPendingReviews({
      course_id: 'software II',
      exam_period: 'spring 2025'
    });

    const tbody = document.querySelector('table tbody');
    tbody.innerHTML = data
      .map(r => `
        <tr>
          <td>${r.course_name}</td>
          <td>${r.exam_period}</td>
          <td>${r.student}</td>
          <td><a class="button" href="/instructor/reply?req=${r.id}">Reply</a></td>
        </tr>
      `)
      .join('');
  } catch (err) {
    flash(err.message);
  }
});
