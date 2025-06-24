// public/js/institution/purchase.js
import { flash } from '../../script.js';
import { purchaseCredits } from '../../api/credits.js';
import { getInstitutions } from '../../api/institution.js';

console.log('🛠️ purchase.js loaded');

async function populateInstitutions() {
  const select = document.querySelector('#inst-name');
  try {
    const list = await getInstitutions();
    // clear placeholder
    select.innerHTML = '<option value="">– choose an institution –</option>';
    list.forEach(inst => {
      const opt = document.createElement('option');
      opt.value = inst.name;
      opt.textContent = inst.name;
      select.appendChild(opt);
    });
  } catch (err) {
    console.error('⚠️ Error loading institutions:', err);
    flash('Could not load institutions');
    // leave the placeholder so user sees no options
  }
}

document.addEventListener('DOMContentLoaded', () => {
  populateInstitutions();

  const form = document.querySelector('#purchase-form');
  form.addEventListener('submit', async e => {
    e.preventDefault();
    console.log('🛠️ submit event fired');

    const instName = form.instName.value;
    const amount   = Number(form.amount.value);
    console.log('🛠️ form values:', { instName, amount });

    try {
      const response = await purchaseCredits({ name: instName, amount });
      console.log('🛠️ API response:', response);
      flash(response.message || 'Purchased successfully!');
    } catch (err) {
      console.error('🛠️ purchaseCredits error:', err);
      flash(err.message || 'Purchase failed');
    }
  });
});
