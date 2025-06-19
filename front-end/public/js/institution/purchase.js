// institution/purchase.js
import { flash } from '../../script.js';
import { purchaseCredits } from '../../api/credits.js';

const form = document.querySelector('#purchase-form');

form.addEventListener('submit', async e => {
  e.preventDefault();
  const amount = Number(form.amount.value);

  try {
    const { data } = await purchaseCredits(amount);
    flash(`Purchased! New balance: ${data.balance}`);
  } catch (err) {
    flash(err.message);
  }
});
