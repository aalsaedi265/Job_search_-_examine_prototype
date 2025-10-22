// Auth token management
export function getAuthToken() {
  return localStorage.getItem('auth_token');
}

export function setAuthToken(token) {
  localStorage.setItem('auth_token', token);
}

export function clearAuth() {
  localStorage.removeItem('auth_token');
  localStorage.removeItem('user');
}

export function setUser(user) {
  localStorage.setItem('user', JSON.stringify(user));
}

export function getUser() {
  const userStr = localStorage.getItem('user');
  return userStr ? JSON.parse(userStr) : null;
}

export function isAuthenticated() {
  return !!getAuthToken();
}
