const API_BASE = '/api/v1';

export async function createProfile(profileData) {
  const response = await fetch(`${API_BASE}/profile`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(profileData)
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to create profile');
  }
  return response.json();
}

export async function uploadResume(userId, file) {
  const formData = new FormData();
  formData.append('resume', file);
  const response = await fetch(`${API_BASE}/profile/${userId}/resume`, {
    method: 'POST',
    body: formData
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to upload resume');
  }
  return response.json();
}

export async function validateProfile(userId) {
  const response = await fetch(`${API_BASE}/profile/${userId}/validate`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to validate profile');
  }
  return response.json();
}

export async function scrapeJobs(keywords, location) {
  const response = await fetch(`${API_BASE}/scrape`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ keywords, location })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to scrape jobs');
  }
  return response.json();
}

export async function getJobs() {
  const response = await fetch(`${API_BASE}/jobs`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get jobs');
  }
  return response.json();
}

export async function applyToJob(jobId, userId) {
  const response = await fetch(`${API_BASE}/apply`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ job_id: jobId, user_id: userId })
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to apply to job');
  }
  return response.json();
}

export async function getApplications(userId) {
  const response = await fetch(`${API_BASE}/applications/${userId}`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get applications');
  }
  return response.json();
}
