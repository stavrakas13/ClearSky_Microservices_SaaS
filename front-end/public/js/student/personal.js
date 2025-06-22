// public/js/student/personal.js
import { flash } from '../../script.js';
import { getPersonalGrades } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params    = new URLSearchParams(location.search);
  const course_id = params.get('course');

  try {
    // endpoint no longer needs exam_period
    const grades = await getPersonalGrades({ course_id });

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
