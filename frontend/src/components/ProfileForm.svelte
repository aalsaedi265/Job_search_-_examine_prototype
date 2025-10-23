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
    color: #ffffff;
    margin-bottom: 1.5rem;
    font-size: 1.1rem;
    font-weight: 600;
    letter-spacing: 1px;
  }

  .user-info strong {
    color: #fce700;
    text-shadow: 0 0 10px rgba(252, 231, 0, 0.5);
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  hr {
    border: none;
    border-top: 2px solid #fce700;
    margin: 2rem 0;
    box-shadow: 0 0 10px rgba(252, 231, 0, 0.3);
  }

  h3 {
    margin-bottom: 1.5rem;
    font-family: 'Teko', sans-serif;
    font-weight: 700;
    font-size: 2rem;
    color: #fce700;
    text-transform: uppercase;
    letter-spacing: 4px;
    text-shadow: 0 0 15px rgba(252, 231, 0, 0.6);
  }

  .work-list {
    margin-bottom: 1.5rem;
  }

  .work-item {
    background: rgba(0, 0, 0, 0.6);
    border: 2px solid #00f0ff;
    border-left: 4px solid #fce700;
    padding: 1.2rem;
    margin-bottom: 1rem;
    box-shadow: 0 0 20px rgba(0, 240, 255, 0.2), inset 0 0 10px rgba(0, 0, 0, 0.5);
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
  }

  .work-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.8rem;
  }

  .work-header strong {
    color: #fce700;
    font-size: 1.2rem;
    font-weight: 700;
    letter-spacing: 1px;
  }

  .work-dates {
    color: #00f0ff;
    font-size: 1rem;
    font-weight: 600;
    letter-spacing: 1px;
  }

  .remove-btn {
    background: #ff003c;
    color: #ffffff;
    border: 2px solid #ff003c;
    width: 32px;
    height: 32px;
    cursor: pointer;
    font-size: 1.5rem;
    font-weight: 900;
    line-height: 1;
    padding: 0;
    transition: all 0.2s;
    clip-path: polygon(4px 0, 100% 0, 100% calc(100% - 4px), calc(100% - 4px) 100%, 0 100%, 0 4px);
    box-shadow: 0 0 15px rgba(255, 0, 60, 0.4);
  }

  .remove-btn:hover {
    background: #ffffff;
    color: #ff003c;
    box-shadow: 0 0 25px rgba(255, 0, 60, 0.8);
    transform: scale(1.1);
  }

  .add-work-section {
    background: rgba(0, 0, 0, 0.4);
    border: 2px solid #555;
    border-left: 3px solid #00f0ff;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.5);
    clip-path: polygon(10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%, 0 10px);
  }

  .add-btn {
    background: #00ff9f;
    color: #000000;
    border: 3px solid #00ff9f;
    padding: 0.8rem 1.5rem;
    cursor: pointer;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1.1rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 2px;
    transition: all 0.2s;
    box-shadow: 0 0 20px rgba(0, 255, 159, 0.4), 0 4px 0 #00aa66;
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
  }

  .add-btn:hover {
    background: #ffffff;
    border-color: #00ff9f;
    box-shadow: 0 0 35px rgba(0, 255, 159, 0.8), 0 2px 0 #00aa66;
    transform: translateY(-3px);
  }

  textarea {
    width: 100%;
    padding: 1rem;
    border: 2px solid #555;
    border-left: 3px solid #fce700;
    background: rgba(0, 0, 0, 0.6);
    color: #ffffff;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1rem;
    font-weight: 500;
    resize: vertical;
    transition: all 0.2s;
  }

  textarea::placeholder {
    color: #666;
  }

  textarea:focus {
    outline: none;
    border-color: #fce700;
    border-left-color: #ff006e;
    box-shadow: 0 0 20px rgba(252, 231, 0, 0.4);
    background: rgba(0, 0, 0, 0.8);
  }

  small {
    display: block;
    margin-top: 0.5rem;
    color: #00f0ff;
    font-size: 0.9rem;
    font-weight: 600;
    letter-spacing: 1px;
  }
</style>
