// student/personal.js
import { flash } from '../../script.js';
import { getPersonalGrades } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  const params      = new URLSearchParams(location.search);
  const course_id   = params.get('course');
  const exam_period = 'spring 2025';
  const user_id     = 'alice';

  try {
    const { data } = await getPersonalGrades({ user_id, course_id, exam_period });

    document.querySelector('input[readonly][value]').value = data.total;
    // populate Q1–Q3 similarly…
  } catch (err) {
    flash(err.message);
  }
});
