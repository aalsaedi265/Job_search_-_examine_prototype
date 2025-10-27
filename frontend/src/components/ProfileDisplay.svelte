<script>
  import { onMount } from 'svelte';
  import { getProfile, deleteProfile } from '../lib/api';
  import { clearAuth } from '../lib/store';

  export let onEdit;
  export let onLogout;

  let profile = null;
  let loading = true;
  let error = '';
  let showDeleteConfirm = false;
  let deleting = false;

  onMount(async () => {
    try {
      profile = await getProfile();
    } catch (err) {
      // If profile doesn't exist, automatically switch to edit mode
      if (err.message.includes('not found') || err.message.includes('404')) {
        handleEdit();
      } else {
        error = err.message;
      }
    } finally {
      loading = false;
    }
  });

  function handleEdit() {
    if (onEdit) onEdit();
  }

  async function handleDelete() {
    if (!showDeleteConfirm) {
      showDeleteConfirm = true;
      return;
    }

    deleting = true;
    try {
      await deleteProfile();
      clearAuth();
      if (onLogout) onLogout();
    } catch (err) {
      error = 'Failed to delete profile: ' + err.message;
      showDeleteConfirm = false;
      deleting = false;
    }
  }

  function cancelDelete() {
    showDeleteConfirm = false;
  }

  function calculateYearsOfExperience() {
    if (!profile?.work_history || profile.work_history.length === 0) return 0;

    let totalYears = 0;
    for (const work of profile.work_history) {
      if (work.start_date) {
        const startDate = new Date(work.start_date);
        const endDate = work.end_date ? new Date(work.end_date) : new Date();
        const years = (endDate - startDate) / (1000 * 60 * 60 * 24 * 365.25);
        totalYears += years;
      }
    }
    return totalYears.toFixed(1);
  }

  function formatDate(dateStr) {
    if (!dateStr) return 'Present';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
  }
</script>

{#if loading}
  <div class="profile-display">
    <div class="loading">Loading profile...</div>
  </div>
{:else if error}
  <div class="profile-display">
    <div class="error">{error}</div>
  </div>
{:else if profile}
  <div class="profile-display">
    <div class="profile-header">
      <div class="header-row-1">
        <h1 class="profile-name">{profile.full_name || 'Anonymous User'}</h1>
        <div class="header-actions">
          <button class="edit-btn" on:click={handleEdit}>Edit Profile</button>
          {#if !showDeleteConfirm}
            <button class="delete-btn" on:click={handleDelete}>Delete Profile</button>
          {/if}
        </div>
      </div>

      <div class="header-row-2">
        <p class="profile-email">{profile.email}</p>
        <div class="experience-badge">{calculateYearsOfExperience()} years of experience</div>
      </div>
    </div>

    {#if showDeleteConfirm}
      <div class="delete-confirm">
        <p class="confirm-message">⚠️ Are you sure you want to delete your profile? This action cannot be undone.</p>
        <div class="confirm-actions">
          <button class="confirm-delete-btn" on:click={handleDelete} disabled={deleting}>
            {deleting ? 'Deleting...' : 'Yes, Delete Profile'}
          </button>
          <button class="cancel-btn" on:click={cancelDelete} disabled={deleting}>Cancel</button>
        </div>
      </div>
    {/if}

    <div class="profile-sections">
      <section class="contact-section">
        <h2 class="section-title">Contact Information</h2>
        <div class="contact-grid">
          <div class="contact-item">
            <span class="contact-label">Phone:</span>
            <span class="contact-value">{profile.phone || 'Not provided'}</span>
          </div>
          <div class="contact-item">
            <span class="contact-label">Location:</span>
            <span class="contact-value">
              {#if profile.address}
                {profile.address.city || ''}{profile.address.city && profile.address.state ? ', ' : ''}{profile.address.state || ''}
                {#if profile.address.zip_code} {profile.address.zip_code}{/if}
              {:else}
                Not provided
              {/if}
            </span>
          </div>
        </div>
      </section>

      {#if profile.work_history && profile.work_history.length > 0}
        <section class="experience-section">
          <h2 class="section-title">Work Experience</h2>
          <div class="experience-list">
            {#each profile.work_history as work}
              <div class="experience-card">
                <div class="exp-header">
                  <div class="exp-title-group">
                    <h3 class="exp-title">{work.title}</h3>
                    <p class="exp-company">{work.company}</p>
                  </div>
                  <div class="exp-dates">
                    {formatDate(work.start_date)} - {formatDate(work.end_date)}
                  </div>
                </div>
                {#if work.description}
                  <p class="exp-description">{work.description}</p>
                {/if}
              </div>
            {/each}
          </div>
        </section>
      {/if}

      {#if profile.skills && profile.skills.length > 0}
        <section class="skills-section">
          <h2 class="section-title">Skills</h2>
          <div class="skills-list">
            {#each profile.skills as skill}
              <span class="skill-tag">{skill}</span>
            {/each}
          </div>
        </section>
      {/if}

      {#if profile.resume_url}
        <section class="resume-section">
          <h2 class="section-title">Resume</h2>
          <a href={profile.resume_url} target="_blank" rel="noopener noreferrer" class="resume-link">
            📄 View Resume (PDF)
          </a>
        </section>
      {/if}
    </div>
  </div>
{/if}

<style>
  .profile-display {
    max-width: 900px;
    margin: 0 auto;
  }

  .loading, .error {
    text-align: center;
    padding: 3rem;
    font-size: 1.3rem;
    color: #fce700;
    font-weight: 600;
    letter-spacing: 2px;
  }

  .error {
    color: #ff003c;
  }

  .profile-header {
    background: rgba(0, 0, 0, 0.7);
    border: 3px solid #fce700;
    border-left: 6px solid #00f0ff;
    padding: 2rem;
    margin-bottom: 2rem;
    box-shadow: 0 0 30px rgba(252, 231, 0, 0.3), inset 0 0 20px rgba(0, 0, 0, 0.5);
    clip-path: polygon(15px 0, 100% 0, 100% calc(100% - 15px), calc(100% - 15px) 100%, 0 100%, 0 15px);
  }

  .header-row-1 {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 2rem;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
  }

  .header-row-2 {
    display: flex;
    align-items: center;
    gap: 2rem;
    flex-wrap: wrap;
  }

  .profile-name {
    font-family: 'Teko', sans-serif;
    font-size: 2.5rem;
    font-weight: 700;
    color: #fce700;
    text-transform: uppercase;
    letter-spacing: 3px;
    text-shadow: 0 0 20px rgba(252, 231, 0, 0.7);
    margin: 0;
    line-height: 1.2;
  }

  .profile-email {
    font-size: 1.1rem;
    color: #00f0ff;
    font-weight: 600;
    letter-spacing: 1px;
    margin: 0;
    line-height: 1.2;
  }

  .experience-badge {
    display: inline-block;
    background: rgba(0, 255, 159, 0.2);
    border: 2px solid #00ff9f;
    color: #00ff9f;
    padding: 0.6rem 1.2rem;
    font-size: 0.9rem;
    font-weight: 700;
    letter-spacing: 1px;
    text-transform: uppercase;
    clip-path: polygon(6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%, 0 6px);
    box-shadow: 0 0 15px rgba(0, 255, 159, 0.3);
    line-height: 1.2;
  }

  .header-actions {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .edit-btn, .delete-btn {
    padding: 0.9rem 1.8rem;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 2px;
    cursor: pointer;
    transition: all 0.2s;
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
  }

  .edit-btn {
    background: #00f0ff;
    color: #000000;
    border: 3px solid #00f0ff;
    box-shadow: 0 0 20px rgba(0, 240, 255, 0.4), 0 4px 0 #0099aa;
  }

  .edit-btn:hover {
    background: #ffffff;
    border-color: #00f0ff;
    box-shadow: 0 0 35px rgba(0, 240, 255, 0.8), 0 2px 0 #0099aa;
    transform: translateY(-3px);
  }

  .delete-btn {
    background: transparent;
    color: #ff003c;
    border: 3px solid #ff003c;
    box-shadow: 0 0 20px rgba(255, 0, 60, 0.2);
  }

  .delete-btn:hover {
    background: #ff003c;
    color: #ffffff;
    box-shadow: 0 0 35px rgba(255, 0, 60, 0.6);
    transform: translateY(-3px);
  }

  .delete-confirm {
    background: rgba(255, 0, 60, 0.15);
    border: 3px solid #ff003c;
    padding: 1.5rem;
    margin-bottom: 2rem;
    clip-path: polygon(10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%, 0 10px);
    box-shadow: 0 0 30px rgba(255, 0, 60, 0.3), inset 0 0 15px rgba(0, 0, 0, 0.5);
  }

  .confirm-message {
    color: #ff003c;
    font-size: 1.1rem;
    font-weight: 700;
    letter-spacing: 1px;
    margin-bottom: 1rem;
  }

  .confirm-actions {
    display: flex;
    gap: 1rem;
  }

  .confirm-delete-btn {
    background: #ff003c;
    color: #ffffff;
    border: 3px solid #ff003c;
    padding: 0.9rem 1.5rem;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 2px;
    cursor: pointer;
    transition: all 0.2s;
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
    box-shadow: 0 0 20px rgba(255, 0, 60, 0.4), 0 4px 0 #aa0028;
  }

  .confirm-delete-btn:hover:not(:disabled) {
    background: #ffffff;
    color: #ff003c;
    box-shadow: 0 0 35px rgba(255, 0, 60, 0.8), 0 2px 0 #aa0028;
    transform: translateY(-3px);
  }

  .confirm-delete-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .cancel-btn {
    background: transparent;
    color: #ffffff;
    border: 3px solid #ffffff;
    padding: 0.9rem 1.5rem;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 2px;
    cursor: pointer;
    transition: all 0.2s;
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
  }

  .cancel-btn:hover:not(:disabled) {
    background: #ffffff;
    color: #000000;
    box-shadow: 0 0 25px rgba(255, 255, 255, 0.6);
  }

  .cancel-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .profile-sections {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .section-title {
    font-family: 'Teko', sans-serif;
    font-size: 1.8rem;
    font-weight: 700;
    color: #fce700;
    text-transform: uppercase;
    letter-spacing: 3px;
    margin: 0 0 1.2rem 0;
    text-shadow: 0 0 15px rgba(252, 231, 0, 0.6);
  }

  .contact-section {
    background: rgba(0, 0, 0, 0.6);
    border: 2px solid #555;
    border-left: 4px solid #00f0ff;
    padding: 1.8rem;
    clip-path: polygon(12px 0, 100% 0, 100% calc(100% - 12px), calc(100% - 12px) 100%, 0 100%, 0 12px);
    box-shadow: 0 0 20px rgba(0, 240, 255, 0.2), inset 0 0 15px rgba(0, 0, 0, 0.5);
  }

  .contact-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.2rem;
  }

  .contact-item {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .contact-label {
    color: #00f0ff;
    font-size: 0.9rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 2px;
  }

  .contact-value {
    color: #ffffff;
    font-size: 1.1rem;
    font-weight: 600;
    letter-spacing: 1px;
  }

  .experience-section {
    background: rgba(0, 0, 0, 0.6);
    border: 2px solid #555;
    border-left: 4px solid #fce700;
    padding: 1.8rem;
    clip-path: polygon(12px 0, 100% 0, 100% calc(100% - 12px), calc(100% - 12px) 100%, 0 100%, 0 12px);
    box-shadow: 0 0 20px rgba(252, 231, 0, 0.2), inset 0 0 15px rgba(0, 0, 0, 0.5);
  }

  .experience-list {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .experience-card {
    background: rgba(0, 0, 0, 0.5);
    border: 2px solid #444;
    border-left: 3px solid #00ff9f;
    padding: 1.5rem;
    clip-path: polygon(10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%, 0 10px);
    box-shadow: inset 0 0 10px rgba(0, 0, 0, 0.5);
  }

  .exp-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 0.8rem;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .exp-title-group {
    flex: 1;
    min-width: 200px;
  }

  .exp-title {
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1.4rem;
    font-weight: 700;
    color: #fce700;
    letter-spacing: 1px;
    margin: 0 0 0.3rem 0;
  }

  .exp-company {
    font-size: 1.1rem;
    color: #00f0ff;
    font-weight: 600;
    letter-spacing: 1px;
    margin: 0;
  }

  .exp-dates {
    color: #00ff9f;
    font-size: 0.95rem;
    font-weight: 700;
    letter-spacing: 1px;
    text-transform: uppercase;
  }

  .exp-description {
    color: #cccccc;
    font-size: 1rem;
    line-height: 1.6;
    margin: 0;
  }

  .skills-section {
    background: rgba(0, 0, 0, 0.6);
    border: 2px solid #555;
    border-left: 4px solid #00ff9f;
    padding: 1.8rem;
    clip-path: polygon(12px 0, 100% 0, 100% calc(100% - 12px), calc(100% - 12px) 100%, 0 100%, 0 12px);
    box-shadow: 0 0 20px rgba(0, 255, 159, 0.2), inset 0 0 15px rgba(0, 0, 0, 0.5);
  }

  .skills-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.8rem;
  }

  .skill-tag {
    background: rgba(0, 240, 255, 0.15);
    border: 2px solid #00f0ff;
    color: #00f0ff;
    padding: 0.6rem 1.2rem;
    font-size: 0.95rem;
    font-weight: 700;
    letter-spacing: 1px;
    clip-path: polygon(6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%, 0 6px);
    box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
    transition: all 0.2s;
  }

  .skill-tag:hover {
    background: rgba(0, 240, 255, 0.3);
    box-shadow: 0 0 20px rgba(0, 240, 255, 0.5);
    transform: translateY(-2px);
  }

  .resume-section {
    background: rgba(0, 0, 0, 0.6);
    border: 2px solid #555;
    border-left: 4px solid #ff003c;
    padding: 1.8rem;
    clip-path: polygon(12px 0, 100% 0, 100% calc(100% - 12px), calc(100% - 12px) 100%, 0 100%, 0 12px);
    box-shadow: 0 0 20px rgba(255, 0, 60, 0.2), inset 0 0 15px rgba(0, 0, 0, 0.5);
  }

  .resume-link {
    display: inline-flex;
    align-items: center;
    gap: 0.8rem;
    background: #fce700;
    color: #000000;
    border: 3px solid #fce700;
    padding: 1rem 2rem;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1.1rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 2px;
    text-decoration: none;
    transition: all 0.2s;
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
    box-shadow: 0 0 20px rgba(252, 231, 0, 0.4), 0 4px 0 #aa9900;
  }

  .resume-link:hover {
    background: #ffffff;
    border-color: #fce700;
    box-shadow: 0 0 35px rgba(252, 231, 0, 0.8), 0 2px 0 #aa9900;
    transform: translateY(-3px);
  }

  @media (max-width: 1200px) {
    .profile-name {
      font-size: 2rem;
    }

    .profile-email {
      font-size: 1rem;
    }

    .experience-badge {
      font-size: 0.85rem;
      padding: 0.5rem 1rem;
    }
  }

  @media (max-width: 768px) {
    .header-row-1 {
      flex-direction: column;
      align-items: flex-start;
      gap: 1rem;
    }

    .header-row-2 {
      flex-direction: column;
      align-items: flex-start;
      gap: 1rem;
    }

    .profile-name {
      font-size: 1.8rem;
    }

    .header-actions {
      width: 100%;
    }

    .edit-btn, .delete-btn {
      flex: 1;
    }

    .contact-grid {
      grid-template-columns: 1fr;
    }

    .exp-header {
      flex-direction: column;
    }
  }
</style>
