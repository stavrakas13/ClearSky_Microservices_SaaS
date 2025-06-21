import { flash } from '../../script.js';
import { purchaseCredits } from '../../api/credits.js';

console.log('🛠️ purchase.js loaded');

const form = document.querySelector('#purchase-form');
if (!form) {
  console.error('⚠️ #purchase-form not found!');
} else {
  form.addEventListener('submit', async e => {
    e.preventDefault();
    console.log('🛠️ submit event fired');

    const instName = form.instName.value.trim();
    const amount = Number(form.amount.value);
    console.log('🛠️ form values:', { instName, amount });

    try {
      const response = await purchaseCredits({ name: instName, amount });
      console.log('🛠️ API response:', response);

      // orchestrator now returns { status, message }
      if (response.message) {
        flash(response.message);
      } else {
        flash('Purchased successfully!');
      }
    } catch (err) {
      console.error('🛠️ purchaseCredits error:', err);
      flash(err.message || 'Purchase failed');
    }
  });
}
