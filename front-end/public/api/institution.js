// institution.js
import { request } from './_request.js';

export const registerInstitution = ({ name, domain, email }) =>
  request('/registration', { method: 'POST', body: { name, domain, email } });

export const uploadInitialGrades = file => {
  const fd = new FormData();
  fd.append('xlsx', file);
  return request('/upload_init', { method: 'POST', body: fd });
};
