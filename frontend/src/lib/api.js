const API_BASE = '/api/v1';

function getAuthHeaders() {
  const token = localStorage.getItem('auth_token');
  return {
    'Content-Type': 'application/json',
    ...(token && { 'Authorization': `Bearer ${token}` })
  };
}

// Auth endpoints
export async function signup(fullName, email, password) {
  const response = await fetch(`${API_BASE}/auth/signup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ full_name: fullName, email, password })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to create account');
  }
  return response.json();
}

export async function login(email, password) {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to login');
  }
  return response.json();
}

export async function getMe() {
  const response = await fetch(`${API_BASE}/auth/me`, {
    headers: getAuthHeaders()
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get user info');
  }
  return response.json();
}

export async function createProfile(profileData) {
  const response = await fetch(`${API_BASE}/profile`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(profileData)
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update profile');
  }
  return response.json();
}

export async function uploadResume(file, shouldParse = false) {
  const token = localStorage.getItem('auth_token');
  const formData = new FormData();
  formData.append('resume', file);
  const url = `${API_BASE}/profile/resume${shouldParse ? '?parse=true' : ''}`;
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      ...(token && { 'Authorization': `Bearer ${token}` })
    },
    body: formData
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to upload resume');
  }
  return response.json();
}

export async function validateProfile() {
  const response = await fetch(`${API_BASE}/profile/validate`, {
    headers: getAuthHeaders()
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to validate profile');
  }
  return response.json();
}

export async function getProfile() {
  const response = await fetch(`${API_BASE}/profile`, {
    headers: getAuthHeaders()
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get profile');
  }
  return response.json();
}

export async function deleteProfile() {
  const response = await fetch(`${API_BASE}/profile`, {
    method: 'DELETE',
    headers: getAuthHeaders()
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete profile');
  }
  return response.json();
}

export async function changePassword(currentPassword, newPassword) {
  const response = await fetch(`${API_BASE}/auth/password`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to change password');
  }
  return response.json();
}

export async function updateEmail(newEmail) {
  const response = await fetch(`${API_BASE}/auth/email`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({ new_email: newEmail })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update email');
  }
  return response.json();
}

export async function scrapeJobs(keywords, location) {
  const response = await fetch(`${API_BASE}/scrape`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ keywords, location })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to scrape jobs');
  }
  return response.json();
}

export async function getJobs() {
  const response = await fetch(`${API_BASE}/jobs`, {
    headers: getAuthHeaders()
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get jobs');
  }
  return response.json();
}

export async function applyToJob(jobId) {
  const response = await fetch(`${API_BASE}/apply`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ job_id: jobId })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to apply to job');
  }
  return response.json();
}

export async function getApplications() {
  const response = await fetch(`${API_BASE}/applications`, {
    headers: getAuthHeaders()
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get applications');
  }
  return response.json();
}
