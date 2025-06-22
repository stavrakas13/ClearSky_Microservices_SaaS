// front-end/public/js/student/personal.js
import { flash } from '../../script.js';
import { getPersonalGrades } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params      = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = params.get('period');       // may be null

  // Visiting /student/personal directly?  Send them back to the course list.
  if (!course_id) {
    flash('Please pick a course first');
    window.location.href = '/student/my-courses';
    return;
  }

  try {
    const grades = await getPersonalGrades({ course_id, exam_period });

    const tbody = document.querySelector('table tbody');
    tbody.innerHTML = Object.entries(grades)
      .map(([component, score]) => `
        <tr>
          <td>${component}</td>
          <td>${score}</td>
        </tr>
      `).join('');
  } catch (err) {
    flash(err.message);
  }
});
