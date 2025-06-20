import { flash } from '../../script.js';
import { purchaseCredits } from '../../api/credits.js';

console.log('🛠️ purchase.js loaded');

const form = document.querySelector('#purchase-form');
console.log('🛠️ purchase form element:', form);

if (!form) {
  console.error('⚠️ #purchase-form not found!');
} else {
  form.addEventListener('submit', async e => {
    e.preventDefault();
    console.log('🛠️ submit event fired');

    const instName = form.instName.value.trim();
    const amount   = Number(form.amount.value);
    console.log('🛠️ form values:', { instName, amount });

    try {
      const response = await purchaseCredits({ name: instName, amount });
      console.log('🛠️ API response:', response);
      const { data } = response;
      flash(`Purchased! New balance: ${data.balance}`);
    } catch (err) {
      console.error('🛠️ purchaseCredits error:', err);
      flash(err.message);
    }
  });
}
