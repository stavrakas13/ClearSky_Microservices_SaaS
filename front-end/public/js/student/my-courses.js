// front-end/public/js/student/my-courses.js
import { flash } from '../../script.js';
import { getStudentCourses } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const courses = await getStudentCourses();      // array from /personal/grades
    const tbody   = document.querySelector('table tbody');

    if (!courses.length) {
      tbody.innerHTML = `
        <tr><td colspan="4" style="text-align:center;">No courses found.</td></tr>
      `;
      return;
    }

    tbody.innerHTML = courses.map(c => {
      // accept camelCase or snake_case
      const courseName  = c.classTitle        ?? c.course_name  ?? '—';
      const examPeriod  = c.declarationPeriod ?? c.exam_period  ?? '—';

      // status might be a string ("open"/"closed") or numeric (0/1)
      let status = c.status ?? c.grading_status;
      if (typeof status === 'number') status = status === 0 ? 'open' : 'closed';
      if (!status) status = '—';

      // course id for link building – fall back to courseName when missing
      const courseId = c.course_id ?? c.id ?? encodeURIComponent(courseName);

      return `
        <tr ${status === 'open' ? 'style="background:#e6e7ea;"' : ''}>
          <td>${courseName}</td>
          <td>${examPeriod}</td>
          <td>${status}</td>
          <td>
            <a href="/student/personal?course=${courseId}&period=${encodeURIComponent(examPeriod)}"
               class="button">View grades</a>

            <a href="/student/request?course=${courseId}&period=${encodeURIComponent(examPeriod)}"
               class="button${status !== 'open' ? ' button--secondary' : ''}">Ask review</a>

            <a href="/student/status?course=${courseId}&period=${encodeURIComponent(examPeriod)}"
               class="button${status === 'open' ? ' button--secondary' : ''}">Status</a>
          </td>
        </tr>`;
    }).join('');
  } catch (err) {
    flash(err.message || 'Failed to load courses');
  }
});
