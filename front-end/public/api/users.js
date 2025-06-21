// file: front-end/public/api/users.js
import { request } from './_request.js';

/**
 * Log in an existing user. On success, returns { role, status, token, userId }.
 */
export const loginUser = ({ username, password, email }) =>
  request('/user/login', {
    method : 'POST',
    body   : email
      ? { email, password }
      : { username, password }
  }).then(response => {
    if (!response.role) {
      throw new Error(response.message || 'Login failed');
    }
    // store JWT if returned
    if (response.token) localStorage.setItem('jwt', response.token);
    return response;
  });
