// student/personal.js
import { flash } from '../../script.js';
import { getPersonalGrades } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params      = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = 'spring 2025'; // or pull from query if dynamic

  try {
    const { data } = await getPersonalGrades({ course_id, exam_period });

    const tbody = document.querySelector('table tbody');
    // data: { total, Q1, Q2, Q3, … }
    tbody.innerHTML = Object.entries(data)
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
