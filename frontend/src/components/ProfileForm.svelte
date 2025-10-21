<script>
  import { createProfile, uploadResume } from '../lib/api';
  import { setUserId, getUserId } from '../lib/store';

  let fullName = '';
  let email = '';
  let phone = '';
  let city = '';
  let state = '';
  let zipCode = '';
  let resumeFile = null;
  let message = '';
  let loading = false;

  async function handleSubmit() {
    loading = true;
    message = '';

    try {
      const profileData = {
        full_name: fullName,
        email: email,
        phone: phone,
        address: {
          city: city,
          state: state,
          zip_code: zipCode
        },
        work_history: [],
        education: [],
        skills: []
      };

      const result = await createProfile(profileData);
      setUserId(result.id);

      if (resumeFile) {
        await uploadResume(result.id, resumeFile);
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

  function handleFileChange(event) {
    resumeFile = event.target.files[0];
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
      <input id="email" type="email" bind:value={email} placeholder="john@example.com" required />
    </div>

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
  .form-row {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr;
    gap: 1rem;
  }

  small {
    display: block;
    margin-top: 0.3rem;
    color: #6b7280;
    font-size: 0.85rem;
  }
</style>
