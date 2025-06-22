// front-end/public/js/student/my-courses.js
import { flash } from '../../script.js';
import { getStudentCourses } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const courses = await getStudentCourses();     // API now returns the array directly

    const tbody = document.querySelector('table tbody');
    tbody.innerHTML = courses
      .map(c => `
        <tr ${c.status === 'open' ? 'style="background:#e6e7ea;"' : ''}>
          <td>${c.course_name}</td>
          <td>${c.exam_period}</td>
          <td>${c.status}</td>
          <td>
            <a href="/student/personal?course=${c.id}&period=${encodeURIComponent(c.exam_period)}" class="button">View grades</a>
            <a href="/student/request?course=${c.id}&period=${encodeURIComponent(c.exam_period)}"  class="button${c.status !== 'open' ? ' button--secondary' : ''}">Ask review</a>
            <a href="/student/status?course=${c.id}&period=${encodeURIComponent(c.exam_period)}"   class="button${c.status === 'open' ? ' button--secondary' : ''}">Status</a>
          </td>
        </tr>
      `)
      .join('');
  } catch (err) {
    flash(err.message);
  }
});
