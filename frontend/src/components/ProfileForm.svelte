<script>
  import { onMount } from 'svelte';
  import { createProfile, uploadResume, getProfile, changePassword, updateEmail } from '../lib/api';
  import { getUser, setAuthToken, setUser } from '../lib/store';

  export let onSaved = null;

  let user = getUser();
  let fullName = '';
  let email = '';
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
  let existingResumeUrl = '';
  let message = '';
  let loading = false;
  let workHistory = [];
  let currentPassword = '';
  let newPassword = '';
  let confirmPassword = '';
  let passwordMessage = '';
  let editingIndex = -1;

  onMount(async () => {
    try {
      const profile = await getProfile();
      console.log('Loaded profile:', profile);
      console.log('Work history from backend:', profile.work_history);
      fullName = profile.full_name || '';
      email = profile.email || user?.email || '';
      phone = profile.phone || '';
      if (profile.address) {
        city = profile.address.city || '';
        state = profile.address.state || '';
        zipCode = profile.address.zip_code || '';
      }
      workHistory = profile.work_history || [];
      console.log('Work history after assignment:', workHistory);
      skills = profile.skills ? profile.skills.join(', ') : '';
      existingResumeUrl = profile.resume_url || '';
    } catch (error) {
      console.error('Error loading profile:', error);
      // Profile not complete yet, that's OK
      email = user?.email || '';
      fullName = user?.name || '';
    }
  });

  async function handleSubmit() {
    loading = true;
    message = '';

    try {
      // Handle email change if email was modified
      const currentUser = getUser();
      if (email && email !== currentUser?.email) {
        try {
          const result = await updateEmail(email);
          if (result.token) {
            setAuthToken(result.token);
            // Update user info in localStorage with new email
            const updatedUser = { ...currentUser, email: email };
            setUser(updatedUser);
            user = updatedUser; // Update local state
          }
          message = 'Email updated successfully. ';
        } catch (error) {
          message = 'Error updating email: ' + error.message + '. ';
        }
      }

      const profileData = {
        full_name: fullName,
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

      console.log('Saving profile with work_history:', workHistory);
      console.log('Full profile data:', profileData);

      await createProfile(profileData);

      if (resumeFile) {
        const uploadResponse = await uploadResume(resumeFile, false);
        console.log('Resume upload response:', uploadResponse);
        message += uploadResponse.message + ' Profile saved successfully!';
        existingResumeUrl = uploadResponse.resume_url;
      } else {
        message += 'Profile saved successfully!';
      }

      // Call the onSaved callback after successful save
      if (onSaved) {
        setTimeout(() => onSaved(), 500); // Small delay to show success message
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

  function editWork(index) {
    editingIndex = index;
  }

  function saveEdit(index) {
    workHistory = [...workHistory];
    editingIndex = -1;
    message = 'Work experience updated! Click Save Profile to persist changes.';
  }

  function cancelEdit() {
    editingIndex = -1;
  }

  function handleFileChange(event) {
    resumeFile = event.target.files[0];
  }

  async function handlePasswordChange() {
    passwordMessage = '';

    // Validation
    if (!currentPassword || !newPassword || !confirmPassword) {
      passwordMessage = 'Error: All password fields are required';
      return;
    }

    if (newPassword !== confirmPassword) {
      passwordMessage = 'Error: New passwords do not match';
      return;
    }

    if (newPassword.length < 6) {
      passwordMessage = 'Error: Password must be at least 6 characters';
      return;
    }

    if (!/[a-zA-Z]/.test(newPassword) || !/[0-9]/.test(newPassword)) {
      passwordMessage = 'Error: Password must contain at least one letter and one number';
      return;
    }

    try {
      await changePassword(currentPassword, newPassword);
      passwordMessage = 'Password changed successfully!';
      // Clear fields
      currentPassword = '';
      newPassword = '';
      confirmPassword = '';
    } catch (error) {
      passwordMessage = 'Error: ' + error.message;
    }
  }

  async function handleEmailChange() {
    const currentUser = getUser();
    if (!email || email === currentUser?.email) {
      message = 'Error: Please enter a different email';
      return;
    }

    try {
      const result = await updateEmail(email);
      // Update the token and user info in localStorage
      if (result.token) {
        setAuthToken(result.token);
        const updatedUser = { ...currentUser, email: email };
        setUser(updatedUser);
        user = updatedUser;
      }
      message = 'Email updated successfully!';
    } catch (error) {
      message = 'Error: ' + error.message;
    }
  }

  // Format date from YYYY-MM-DD to MM/YYYY
  function formatDate(dateStr) {
    if (!dateStr) return 'Present';

    // Handle YYYY-MM-DD format
    if (dateStr.includes('-')) {
      const parts = dateStr.split('-');
      // parts[0] = YYYY, parts[1] = MM, parts[2] = DD
      // Remove leading zero from month if present
      const month = parts[1].startsWith('0') ? parts[1].substring(1) : parts[1];
      return `${month}/${parts[0]}`;
    }

    // Already in correct format or other format
    return dateStr;
  }
</script>

<div class="form-container">
  <h2>Your Profile</h2>

  <form on:submit|preventDefault={handleSubmit}>
    <div class="form-group">
      <label for="fullName">Full Name *</label>
      <input id="fullName" type="text" bind:value={fullName} placeholder="John Doe" required />
    </div>

    <div class="form-group">
      <label for="email">Email *</label>
      <input id="email" type="email" bind:value={email} placeholder="you@example.com" required />
      <small>Changing your email will require a new authentication token</small>
    </div>

    <hr />

    <h3>Change Password (Optional)</h3>
    <div class="password-section">
      <div class="form-group">
        <label for="currentPassword">Current Password</label>
        <input id="currentPassword" type="password" bind:value={currentPassword} placeholder="Enter current password" />
      </div>

      <div class="form-row">
        <div class="form-group">
          <label for="newPassword">New Password</label>
          <input id="newPassword" type="password" bind:value={newPassword} placeholder="At least 6 characters" />
        </div>
        <div class="form-group">
          <label for="confirmPassword">Confirm New Password</label>
          <input id="confirmPassword" type="password" bind:value={confirmPassword} placeholder="Re-enter new password" />
        </div>
      </div>

      <button type="button" class="password-btn" on:click={handlePasswordChange}>
        Change Password
      </button>

      {#if passwordMessage}
        <p class="message" class:error={passwordMessage.startsWith('Error')}>{passwordMessage}</p>
      {/if}
    </div>

    <hr />

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

    <div class="form-group">
      <label for="skills">Skills (comma-separated)</label>
      <input id="skills" type="text" bind:value={skills} placeholder="JavaScript, Python, React" />
    </div>

    <div class="form-group">
      <label for="resume">Resume (PDF)</label>
      {#if existingResumeUrl}
        <div class="resume-status">
          <span class="resume-indicator">✓ Resume uploaded</span>
          <a href={existingResumeUrl} target="_blank" rel="noopener noreferrer" class="view-resume">View Current</a>
        </div>
      {/if}
      <input id="resume" type="file" accept=".pdf" on:change={handleFileChange} />
      <small>{existingResumeUrl ? 'Replace your resume by selecting a new PDF file' : 'Upload your resume in PDF format'}</small>
    </div>

    <hr />

    <h3>Work History (Required for Job Matching)</h3>

    {#if workHistory.length > 0}
      <div class="work-list">
        {#each workHistory as work, index}
          <div class="work-item">
            {#if editingIndex === index}
              <div class="edit-mode">
                <div class="form-row">
                  <div class="form-group">
                    <label>Company</label>
                    <input type="text" bind:value={work.company} placeholder="Acme Corp" />
                  </div>
                  <div class="form-group">
                    <label>Job Title</label>
                    <input type="text" bind:value={work.title} placeholder="Software Engineer" />
                  </div>
                </div>
                <div class="form-row">
                  <div class="form-group">
                    <label>Start Date</label>
                    <input type="date" bind:value={work.start_date} />
                  </div>
                  <div class="form-group">
                    <label>End Date (leave empty if current)</label>
                    <input type="date" bind:value={work.end_date} />
                  </div>
                </div>
                <div class="form-group">
                  <label>Description</label>
                  <textarea bind:value={work.description} placeholder="Brief description..." rows="2"></textarea>
                </div>
                <div class="edit-actions">
                  <button type="button" class="save-edit-btn" on:click={() => saveEdit(index)}>Save</button>
                  <button type="button" class="cancel-edit-btn" on:click={cancelEdit}>Cancel</button>
                </div>
              </div>
            {:else}
              <div class="work-header">
                <div class="work-title-line">
                  {#if work.company && work.title}
                    <strong>{work.company}</strong> | {work.title}
                  {:else if work.company}
                    <strong>{work.company}</strong>
                  {:else if work.title}
                    <strong>{work.title}</strong>
                  {:else}
                    <strong>Work Experience</strong>
                  {/if}
                </div>
                <div class="work-actions">
                  <button type="button" class="edit-btn" on:click={() => editWork(index)}>✎</button>
                  <button type="button" class="remove-btn" on:click={() => removeWork(index)}>×</button>
                </div>
              </div>
              <div class="work-dates">
                {formatDate(work.start_date)} - {formatDate(work.end_date)}
              </div>
              {#if work.description}
                <div class="work-description">{@html work.description.replace(/\n/g, '<br>')}</div>
              {/if}
            {/if}
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

    <button type="submit" disabled={loading}>
      {loading ? 'Saving...' : 'Save Profile'}
    </button>

    {#if message}
      <p class="message" class:error={message.startsWith('Error')}>{message}</p>
    {/if}
  </form>
</div>

<style>
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

  .work-description {
    margin-top: 0.8rem;
    color: #cccccc;
    font-size: 0.95rem;
    line-height: 1.6;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .work-title-line {
    flex: 1;
  }

  .work-title-line strong {
    color: #fce700;
  }

  .work-actions {
    display: flex;
    gap: 0.5rem;
  }

  .edit-btn {
    background: #00f0ff;
    color: #000000;
    border: 2px solid #00f0ff;
    width: 32px;
    height: 32px;
    cursor: pointer;
    font-size: 1.2rem;
    font-weight: 900;
    line-height: 1;
    padding: 0;
    transition: all 0.2s;
    clip-path: polygon(4px 0, 100% 0, 100% calc(100% - 4px), calc(100% - 4px) 100%, 0 100%, 0 4px);
    box-shadow: 0 0 15px rgba(0, 240, 255, 0.4);
  }

  .edit-btn:hover {
    background: #ffffff;
    color: #00f0ff;
    box-shadow: 0 0 25px rgba(0, 240, 255, 0.8);
    transform: scale(1.1);
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

  .edit-mode {
    padding: 0.5rem 0;
  }

  .edit-mode .form-group {
    margin-bottom: 1rem;
  }

  .edit-actions {
    display: flex;
    gap: 0.8rem;
    margin-top: 1rem;
  }

  .save-edit-btn {
    background: #00ff9f;
    color: #000000;
    border: 2px solid #00ff9f;
    padding: 0.6rem 1.2rem;
    cursor: pointer;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 1px;
    transition: all 0.2s;
    clip-path: polygon(6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%, 0 6px);
    box-shadow: 0 0 15px rgba(0, 255, 159, 0.4);
  }

  .save-edit-btn:hover {
    background: #ffffff;
    color: #00ff9f;
    box-shadow: 0 0 25px rgba(0, 255, 159, 0.8);
  }

  .cancel-edit-btn {
    background: transparent;
    color: #ff003c;
    border: 2px solid #ff003c;
    padding: 0.6rem 1.2rem;
    cursor: pointer;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 1px;
    transition: all 0.2s;
    clip-path: polygon(6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%, 0 6px);
  }

  .cancel-edit-btn:hover {
    background: #ff003c;
    color: #ffffff;
    box-shadow: 0 0 20px rgba(255, 0, 60, 0.6);
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

  input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    background: rgba(0, 0, 0, 0.4);
    border-color: #444;
  }

  .resume-status {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 0.8rem;
    padding: 0.8rem;
    background: rgba(0, 255, 159, 0.1);
    border: 2px solid #00ff9f;
    border-radius: 4px;
  }

  .resume-indicator {
    color: #00ff9f;
    font-weight: 700;
    letter-spacing: 1px;
  }

  .view-resume {
    color: #00f0ff;
    text-decoration: none;
    font-weight: 700;
    letter-spacing: 1px;
    padding: 0.4rem 1rem;
    border: 2px solid #00f0ff;
    transition: all 0.2s;
    clip-path: polygon(4px 0, 100% 0, 100% calc(100% - 4px), calc(100% - 4px) 100%, 0 100%, 0 4px);
  }

  .view-resume:hover {
    background: #00f0ff;
    color: #000000;
    box-shadow: 0 0 15px rgba(0, 240, 255, 0.5);
  }

  .password-section {
    background: rgba(0, 0, 0, 0.4);
    border: 2px solid #555;
    border-left: 3px solid #ff003c;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.5);
    clip-path: polygon(10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%, 0 10px);
  }

  .password-btn {
    background: #ff003c;
    color: #ffffff;
    border: 3px solid #ff003c;
    padding: 0.8rem 1.5rem;
    cursor: pointer;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1.1rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 2px;
    transition: all 0.2s;
    box-shadow: 0 0 20px rgba(255, 0, 60, 0.4), 0 4px 0 #aa0028;
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
  }

  .password-btn:hover {
    background: #ffffff;
    color: #ff003c;
    border-color: #ff003c;
    box-shadow: 0 0 35px rgba(255, 0, 60, 0.8), 0 2px 0 #aa0028;
    transform: translateY(-3px);
  }

</style>
