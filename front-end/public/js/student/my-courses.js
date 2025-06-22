// student/my-courses.js
import { flash } from '../../script.js';
import { getStudentCourses } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    // now GET /stats/available
    const { data } = await getStudentCourses();

    // Assuming data is an array of submissions: { course_name, exam_period, status, id }
    const tbody = document.querySelector('table tbody');
    tbody.innerHTML = data
      .map(c => `
        <tr ${c.status === 'open' ? 'style="background:#e6e7ea;"' : ''}>
          <td>${c.course_name}</td>
          <td>${c.exam_period}</td>
          <td>${c.status}</td>
          <td>
            <a href="/student/personal?course=${c.id}" class="button">View grades</a>
            <a href="/student/request?course=${c.id}" class="button${c.status !== 'open' ? ' button--secondary' : ''}">Ask review</a>
            <a href="/student/status?course=${c.id}"  class="button${c.status === 'open' ? ' button--secondary' : ''}">Status</a>
          </td>
        </tr>
      `)
      .join('');
  } catch (err) {
    flash(err.message);
  }
});
