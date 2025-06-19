// student/statistics.js
import { flash } from '../../script.js';
import { getDistributions } from '../../api/stats.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const { data } = await getDistributions({ /* filters */ });
    console.log('Stats:', data);
    flash('Statistics loaded—see console.');
  } catch (err) {
    flash(err.message);
  }
});
