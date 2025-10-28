<script>
  import { onMount } from 'svelte';
  import { scrapeJobs, validateProfile } from '../lib/api';

  let keywords = '';
  let location = '';
  let message = '';
  let loading = false;
  let profileValid = false;
  let yearsOfExperience = 0;
  let validationMessage = '';
  let checkingProfile = true;

  onMount(async () => {
    try {
      const validation = await validateProfile();
      console.log('Validation response:', validation);
      profileValid = validation.is_complete;
      yearsOfExperience = validation.years_of_experience || 0;
      validationMessage = validation.message;

      // Show missing fields in validation message for debugging
      if (!validation.is_complete && validation.missing_fields) {
        validationMessage = `Missing fields: ${validation.missing_fields.join(', ')}. ${validation.message}`;
      }
    } catch (error) {
      console.error('Validation error:', error);
      validationMessage = 'Unable to validate profile. Please complete your profile first.';
    }
    checkingProfile = false;
  });

  async function handleScrape() {
    if (!profileValid) {
      message = 'Please complete your profile before searching for jobs';
      return;
    }

    if (!keywords || !location) {
      message = 'Please enter keywords and location';
      return;
    }

    loading = true;
    message = 'Scraping jobs from Indeed...';

    try {
      const result = await scrapeJobs(keywords, location);
      message = `Found ${result.jobs_scraped || 0} new jobs! Check the Jobs tab.`;

      // Clear form after successful scrape
      setTimeout(() => {
        keywords = '';
        location = '';
      }, 2000);
    } catch (error) {
      message = 'Error: ' + error.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="form-container">
  <h2>Search for Jobs</h2>

  {#if checkingProfile}
    <p class="description">Checking profile...</p>
  {:else if !profileValid}
    <div class="warning-box">
      <strong>⚠️ Profile Incomplete</strong>
      <p>{validationMessage}</p>
      <p>Please complete your profile in the Profile tab before searching for jobs. Make sure to include your work history to calculate years of experience.</p>
    </div>
  {:else}
    <div class="info-box">
      <strong>✓ Profile Complete</strong>
      <p>Years of Experience: <strong>{yearsOfExperience.toFixed(1)}</strong> years</p>
    </div>
  {/if}

  <p class="description">
    Scrape job listings from Indeed based on your search criteria.
  </p>

  <div class="form-group">
    <label for="keywords">Keywords *</label>
    <input
      id="keywords"
      type="text"
      bind:value={keywords}
      placeholder="e.g. Software Engineer"
      disabled={loading || !profileValid}
    />
  </div>

  <div class="form-group">
    <label for="location">Location *</label>
    <input
      id="location"
      type="text"
      bind:value={location}
      placeholder="e.g. San Francisco, CA"
      disabled={loading || !profileValid}
    />
  </div>

  <button on:click={handleScrape} disabled={loading || !keywords || !location || !profileValid}>
    {loading ? 'Scraping...' : 'Scrape Indeed Jobs'}
  </button>

  {#if message}
    <p class="message" class:error={message.startsWith('Error')}>{message}</p>
  {/if}
</div>

<style>
  .description {
    color: #6b7280;
    margin-bottom: 1.5rem;
  }

  .warning-box {
    background-color: #fef3c7;
    border: 2px solid #f59e0b;
    border-radius: 8px;
    padding: 1rem;
    margin-bottom: 1.5rem;
  }

  .warning-box strong {
    color: #d97706;
    display: block;
    margin-bottom: 0.5rem;
  }

  .warning-box p {
    color: #92400e;
    margin: 0.5rem 0;
    font-size: 0.9rem;
  }

  .info-box {
    background-color: #d1fae5;
    border: 2px solid #10b981;
    border-radius: 8px;
    padding: 1rem;
    margin-bottom: 1.5rem;
  }

  .info-box strong {
    color: #047857;
    display: block;
    margin-bottom: 0.5rem;
  }

  .info-box p {
    color: #065f46;
    margin: 0.5rem 0;
    font-size: 0.9rem;
  }
</style>
