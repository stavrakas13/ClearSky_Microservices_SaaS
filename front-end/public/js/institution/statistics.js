// public/js/institution/statistics.js
import { flash } from '../../script.js';
import { getAvailableStats } from '../../api/stats.js';

const btn = document.querySelector('button.button--secondary');

btn.addEventListener('click', async () => {
  try {
    const stats = await getAvailableStats();
    console.log('Available stats:', stats);
    flash('Available stats fetched—check console.');
  } catch (err) {
    flash(err.message);
  }
});
