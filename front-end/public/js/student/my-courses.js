// public/js/student/my-courses.js
import { flash } from '../script.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const res = await fetch('/personal/courses', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: 'alice' }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    const tbody = document.querySelector('table tbody');
    tbody.innerHTML = body.data.map(c => `
      <tr ${c.status==='open'?'style="background:#e6e7ea;"':''}>
        <td>${c.course_name}</td>
        <td>${c.exam_period}</td>
        <td>${c.status}</td>
        <td>
          <a href="/student/personal?course=${c.id}" class="button">View grades</a>
          <a href="/student/request?course=${c.id}" class="button${c.status!=='open'?' button--secondary':''}">Ask review</a>
          <a href="/student/status?course=${c.id}"  class="button${c.status==='open'?' button--secondary':''}">Status</a>
        </td>
      </tr>
    `).join('');
  } catch (err) {
    flash(err.message);
  }
});
