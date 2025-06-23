// front-end/public/js/student/personal.js
import { flash } from '../../script.js';
import { getPersonalGrades } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params      = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = params.get('period');

  if (!course_id) {
    flash('Please pick a course first');
    location.href = '/student/my-courses';
    return;
  }

  try {
    const grades = await getPersonalGrades({ course_id, exam_period });
    const tbody  = document.querySelector('table tbody');

    if (!grades.length) {
      tbody.innerHTML = `
        <tr>
          <td colspan="4" style="text-align:center;">No grades found.</td>
        </tr>`;
      return;
    }

    tbody.innerHTML = grades.map(g => {
      const period = g.declarationPeriod ?? g.declaration_period ?? '—';
      const course = g.classTitle        ?? g.class_title        ?? '—';
      const status = g.gradingStatus     ?? g.grading_status     ?? '—';
      const score  = g.grade             ?? g.score              ?? '—';

      return `
        <tr>
          <td>${period}</td>
          <td>${course}</td>
          <td>${status}</td>
          <td>${score}</td>
        </tr>`;
    }).join('');
  } catch (err) {
    flash(err.message || 'Failed to load grades');
  }
});
