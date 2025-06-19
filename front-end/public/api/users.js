// users.js
import { request } from './_request.js';

export const registerUser = payload =>
  request('/user/register', { method: 'POST', body: payload });

export const loginUser = ({ username, password }) =>
  request('/user/login',   { method: 'POST', body: { username, password } });

export const deleteUser = ({ user_id }) =>
  request('/user/delete',  { method: 'DELETE', body: { user_id } });

export const googleLoginUser = token =>
  request('/user/google-login', { method: 'POST', body: { token } });
