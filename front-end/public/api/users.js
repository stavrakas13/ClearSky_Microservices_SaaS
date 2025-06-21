// front-end/public/api/users.js
import { request } from './_request.js';

/**
 * Register a new user (no change needed here)
 */
export const registerUser = payload =>
  request('/user/register', { method: 'POST', body: payload });

/**
 * Log in an existing user. On success, store the JWT.
 * Expects backend to return { data: { token: "…" } }
 */
export const loginUser = ({ username, password }) =>
  request('/user/login', { method: 'POST', body: { username, password } })
    .then(({ data }) => {
      localStorage.setItem('jwt', data.token);
      return data;
    });

/**
 * Delete a user (no change needed)
 */
export const deleteUser = ({ user_id }) =>
  request('/user/delete', { method: 'DELETE', body: { user_id } });

/**
 * Google login (no change needed)
 */
export const googleLoginUser = token =>
  request('/user/google-login', { method: 'POST', body: { token } });
