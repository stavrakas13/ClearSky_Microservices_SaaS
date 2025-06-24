// public/js/api/institution.js
import { request } from './_request.js';

export const registerInstitution = ({ name, email, director }) =>
  request('/registration', {
    method: 'POST',
    body: { name, email, director }
  });

export const uploadInitialGrades = file => {
  const fd = new FormData();
  fd.append('xlsx', file);
  return request('/upload_init', { method: 'POST', body: fd });
};

// ← new helper:
export const getInstitutions = () =>
  request('/institutions', { method: 'GET' });
