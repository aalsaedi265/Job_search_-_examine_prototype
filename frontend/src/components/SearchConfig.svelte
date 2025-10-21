<script>
  import { scrapeJobs } from '../lib/api';

  let keywords = '';
  let location = '';
  let message = '';
  let loading = false;

  async function handleScrape() {
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
      disabled={loading}
    />
  </div>

  <div class="form-group">
    <label for="location">Location *</label>
    <input
      id="location"
      type="text"
      bind:value={location}
      placeholder="e.g. San Francisco, CA"
      disabled={loading}
    />
  </div>

  <button on:click={handleScrape} disabled={loading || !keywords || !location}>
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
</style>
