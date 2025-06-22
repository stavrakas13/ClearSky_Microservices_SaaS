// front-end/public/api/users.js
import { request } from './_request.js';

/**
 * Register a new user by username/password/role only.
 * @param {{ username: string, password: string, role: string }} payload
 */
export const registerUser = ({ id, password, role }) =>
  request('/user/register', {
    method: 'POST',
    body: { username: id, password, role }
  }).then(response => {
    if (response.error) throw new Error(response.error);
    return response;
  });

/**
 * Log in an existing user.
 */
export const loginUser = ({ username, password }) =>
  request('/user/login', {
    method: 'POST',
    body: { username, password }
  }).then(response => {
    if (!response.role) throw new Error(response.message || 'Login failed');
    if (response.token) localStorage.setItem('jwt', response.token);
    return response;
  });
