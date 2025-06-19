import { flash } from '../../script.js';

const form = document.querySelector('#purchase-form');
form.addEventListener('submit', async e => {
  e.preventDefault();

  const instName = form.instName.value;
  const amount   = Number(form.amount.value);

  try {
    const res = await fetch('/api/purchase', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: instName, amount }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    flash(`Purchased ${amount} credits for ${instName}! New balance: ${body.data.balance}`);
  } catch (err) {
    flash(err.message);
  }
});
