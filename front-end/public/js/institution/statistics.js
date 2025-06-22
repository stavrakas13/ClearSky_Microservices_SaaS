// public/js/institution/statistics.js
import { flash } from '../../script.js';
import { getDistributions } from '../../api/stats.js';

const btn = document.querySelector('button.button--secondary');

btn.addEventListener('click', async () => {
  try {
    const stats = await getDistributions({ /* filters */ });
    console.log('Distributions:', stats);
    flash('Distributions fetched—check console.');
  } catch (err) {
    flash(err.message);
  }
});
