// file: front-end/public/api/users.js
import { request } from './_request.js';

/**
 * Register a new user
 */
export const registerUser = payload =>
  request('/user/register', { method: 'POST', body: payload });

/**
 * Log in an existing user. On success, store the JWT.
 * Expects backend to return { role, status, token, userId }
 */
export const loginUser = ({ username, password }) =>
  request('/user/login', { method: 'POST', body: { username, password } })
    .then(response => {
      // Store the raw token from the top-level response
      localStorage.setItem('jwt', response.token);
      return response;
    });

/**
 * Delete a user
 */
export const deleteUser = ({ user_id }) =>
  request('/user/delete', { method: 'DELETE', body: { user_id } });

/**
 * Google login
 */
export const googleLoginUser = token =>
  request('/user/google-login', { method: 'POST', body: { token } });
