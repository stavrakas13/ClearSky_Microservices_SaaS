// public/js/instructor/review-list.js
import { flash } from '../script.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const res = await fetch('/instructor/review-list', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ course_id: 'software II', exam_period: 'spring 2025' }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    const tbody = document.querySelector('table tbody');
    tbody.innerHTML = body.data.map(r => `
      <tr>
        <td>${r.course_name}</td>
        <td>${r.exam_period}</td>
        <td>${r.student}</td>
        <td><a class="button" href="/instructor/reply?req=${r.id}">Reply</a></td>
      </tr>
    `).join('');
  } catch (err) {
    flash(err.message);
  }
});
