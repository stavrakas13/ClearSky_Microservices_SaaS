import { flash } from '../../script.js';
import { purchaseCredits } from '../../api/credits.js';

const form = document.querySelector('#purchase-form');
form.addEventListener('submit', async e => {
  e.preventDefault();

  // Διάβασε το instName και απόφυγε κενά
  const instName = form.instName.value.trim();
  const amount   = Number(form.amount.value);

  try {
    // Στείλε το σωστό πεδίο name στο backend
    const { data } = await purchaseCredits({ name: instName, amount });
    flash(`Purchased! New balance: ${data.balance}`);
  } catch (err) {
    flash(err.message);
  }
});
