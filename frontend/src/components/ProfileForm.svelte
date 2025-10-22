<script>
  import { onMount } from 'svelte';
  import { createProfile, uploadResume, getProfile } from '../lib/api';
  import { getUser } from '../lib/store';

  let user = getUser();
  let phone = '';
  let city = '';
  let state = '';
  let zipCode = '';
  let company = '';
  let title = '';
  let startDate = '';
  let endDate = '';
  let workDescription = '';
  let skills = '';
  let resumeFile = null;
  let message = '';
  let loading = false;
  let workHistory = [];

  onMount(async () => {
    try {
      const profile = await getProfile();
      phone = profile.phone || '';
      if (profile.address) {
        city = profile.address.city || '';
        state = profile.address.state || '';
        zipCode = profile.address.zip_code || '';
      }
      workHistory = profile.work_history || [];
      skills = profile.skills ? profile.skills.join(', ') : '';
    } catch (error) {
      // Profile not complete yet, that's OK
    }
  });

  async function handleSubmit() {
    loading = true;
    message = '';

    try {
      const profileData = {
        phone: phone,
        address: {
          city: city,
          state: state,
          zip_code: zipCode
        },
        work_history: workHistory,
        education: [],
        skills: skills.split(',').map(s => s.trim()).filter(s => s)
      };

      await createProfile(profileData);

      if (resumeFile) {
        await uploadResume(resumeFile);
        message = 'Profile and resume saved successfully!';
      } else {
        message = 'Profile saved successfully!';
      }

    } catch (error) {
      message = 'Error: ' + error.message;
    } finally {
      loading = false;
    }
  }

  function addWorkHistory() {
    if (!company || !title || !startDate) {
      message = 'Please fill company, title, and start date';
      return;
    }

    workHistory = [...workHistory, {
      company,
      title,
      start_date: startDate,
      end_date: endDate || '',
      description: workDescription
    }];

    // Clear form
    company = '';
    title = '';
    startDate = '';
    endDate = '';
    workDescription = '';
    message = 'Work experience added! Click Save Profile to persist.';
  }

  function removeWork(index) {
    workHistory = workHistory.filter((_, i) => i !== index);
  }

  function handleFileChange(event) {
    resumeFile = event.target.files[0];
  }
</script>

<div class="form-container">
  <h2>Your Profile</h2>
  <p class="user-info">Logged in as: <strong>{user?.email}</strong></p>

  <form on:submit|preventDefault={handleSubmit}>
    <div class="form-group">
      <label for="phone">Phone *</label>
      <input id="phone" type="tel" bind:value={phone} placeholder="555-123-4567" required />
    </div>

    <div class="form-row">
      <div class="form-group">
        <label for="city">City *</label>
        <input id="city" type="text" bind:value={city} placeholder="San Francisco" required />
      </div>

      <div class="form-group">
        <label for="state">State *</label>
        <input id="state" type="text" bind:value={state} placeholder="CA" required />
      </div>

      <div class="form-group">
        <label for="zipCode">Zip Code</label>
        <input id="zipCode" type="text" bind:value={zipCode} placeholder="94102" />
      </div>
    </div>

    <hr />

    <h3>Work History (Required for Job Matching)</h3>

    {#if workHistory.length > 0}
      <div class="work-list">
        {#each workHistory as work, index}
          <div class="work-item">
            <div class="work-header">
              <strong>{work.title}</strong> at {work.company}
              <button type="button" class="remove-btn" on:click={() => removeWork(index)}>×</button>
            </div>
            <div class="work-dates">
              {work.start_date} - {work.end_date || 'Present'}
            </div>
          </div>
        {/each}
      </div>
    {/if}

    <div class="add-work-section">
      <div class="form-row">
        <div class="form-group">
          <label for="company">Company</label>
          <input id="company" type="text" bind:value={company} placeholder="Acme Corp" />
        </div>
        <div class="form-group">
          <label for="title">Job Title</label>
          <input id="title" type="text" bind:value={title} placeholder="Software Engineer" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label for="startDate">Start Date</label>
          <input id="startDate" type="date" bind:value={startDate} />
        </div>
        <div class="form-group">
          <label for="endDate">End Date (leave empty if current)</label>
          <input id="endDate" type="date" bind:value={endDate} />
        </div>
      </div>

      <div class="form-group">
        <label for="workDescription">Description</label>
        <textarea id="workDescription" bind:value={workDescription} placeholder="Brief description of responsibilities..." rows="2"></textarea>
      </div>

      <button type="button" class="add-btn" on:click={addWorkHistory}>+ Add Work Experience</button>
    </div>

    <hr />

    <div class="form-group">
      <label for="skills">Skills (comma-separated)</label>
      <input id="skills" type="text" bind:value={skills} placeholder="JavaScript, Python, React" />
    </div>

    <div class="form-group">
      <label for="resume">Resume (PDF)</label>
      <input id="resume" type="file" accept=".pdf" on:change={handleFileChange} />
      <small>Upload your resume in PDF format</small>
    </div>

    <button type="submit" disabled={loading}>
      {loading ? 'Saving...' : 'Save Profile'}
    </button>

    {#if message}
      <p class="message" class:error={message.startsWith('Error')}>{message}</p>
    {/if}
  </form>
</div>

<style>
  .user-info {
    color: #6b7280;
    margin-bottom: 1.5rem;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  hr {
    border: none;
    border-top: 1px solid #e5e7eb;
    margin: 2rem 0;
  }

  h3 {
    margin-bottom: 1rem;
    color: #1f2937;
  }

  .work-list {
    margin-bottom: 1rem;
  }

  .work-item {
    background: #f3f4f6;
    padding: 1rem;
    border-radius: 6px;
    margin-bottom: 0.5rem;
  }

  .work-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .work-dates {
    color: #6b7280;
    font-size: 0.9rem;
  }

  .remove-btn {
    background: #ef4444;
    color: white;
    border: none;
    border-radius: 50%;
    width: 24px;
    height: 24px;
    cursor: pointer;
    font-size: 1.2rem;
    line-height: 1;
    padding: 0;
  }

  .remove-btn:hover {
    background: #dc2626;
  }

  .add-work-section {
    background: #f9fafb;
    padding: 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }

  .add-btn {
    background: #10b981;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
  }

  .add-btn:hover {
    background: #059669;
  }

  textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-family: inherit;
    resize: vertical;
  }

  small {
    display: block;
    margin-top: 0.3rem;
    color: #6b7280;
    font-size: 0.85rem;
  }
</style>
