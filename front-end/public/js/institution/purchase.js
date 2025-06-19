import { flash } from '../script.js';

const form = document.querySelector('fieldset form');
form.addEventListener('submit', async e => {
  e.preventDefault();
  const amount = Number(form.amount.value);

  try {
    const res = await fetch('/purchase', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: "NTUA",
        amount: 5
      }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.message || res.statusText);

    flash(`Purchased! New balance: ${body.data.balance}`);
  } catch (err) {
    flash(err.message);
  }
});
