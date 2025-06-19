// public/js/institution/statistics.js
import { flash } from '../script.js';

const btn = document.querySelector('button.button--secondary');
btn.addEventListener('click', async () => {
  try {
    const res = await fetch('/stats/distributions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ /* add your filters here */ }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    console.log('Distributions:', body.data);
    flash('Distributions fetched—check console.');
  } catch (err) {
    flash(err.message);
  }
});
