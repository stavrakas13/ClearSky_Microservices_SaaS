// public/js/student/statistics.js
import { flash } from '../script.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const res = await fetch('/stats/distributions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ /* filters */ }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    console.log('Stats:', body.data);
    flash('Statistics loaded—see console.');
  } catch (err) {
    flash(err.message);
  }
});
