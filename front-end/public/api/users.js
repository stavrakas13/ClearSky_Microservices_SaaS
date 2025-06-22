import { request } from './_request.js';

/**
 * Register a new user.
 * @param {{ username: string, password: string, role: string, student_id?: string }} payload
 */
export const registerUser = ({ username, password, role, student_id }) =>
  request('/user/register', {
    method: 'POST',
    body  : { username, password, role, student_id }
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
    body  : { username, password }
  }).then(response => {
    if (!response.role) throw new Error(response.message || 'Login failed');
    if (response.token) localStorage.setItem('jwt', response.token);
    return response;
  });

/**
 * Change password for an existing user.
 * @param {{ username: string, old_password: string, new_password: string }} payload
 */
export const changePassword = ({ username, old_password, new_password }) =>
  request('/user/change-password', {
    method: 'PATCH',
    body  : { username, old_password, new_password }
  });

/**
 * Login via Google token.
 * @param {string} token  Google ID token
 */
export const googleLoginUser = token =>
  request('/user/google-login', {
    method: 'POST',
    body  : { token }
  });
