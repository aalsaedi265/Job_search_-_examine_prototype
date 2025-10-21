export function getUserId() {
  return localStorage.getItem('user_id');
}

export function setUserId(id) {
  localStorage.setItem('user_id', id);
}

export function hasUserId() {
  return !!getUserId();
}
