<script>
  import { onMount } from 'svelte';
  import { getJobs, applyToJob } from '../lib/api';

  let jobs = [];
  let loading = false;
  let applyingTo = null;
  let errorMessage = '';

  onMount(async () => {
    await loadJobs();
  });

  async function loadJobs() {
    loading = true;
    errorMessage = '';
    try {
      jobs = await getJobs();
    } catch (error) {
      errorMessage = 'Error loading jobs: ' + error.message;
      jobs = [];
    } finally {
      loading = false;
    }
  }

  async function handleApply(jobId, jobTitle) {
    const confirmed = confirm(
      `Apply to this job?\n\n"${jobTitle}"\n\nThis will open a browser window and auto-fill the application form.`
    );

    if (!confirmed) return;

    applyingTo = jobId;
    try {
      const result = await applyToJob(jobId);

      let resultMessage = `Application Status: ${result.status}\n\n${result.message}`;

      if (result.fields_filled && result.fields_filled.length > 0) {
        resultMessage += `\n\nFields filled: ${result.fields_filled.join(', ')}`;
      }

      if (result.errors && result.errors.length > 0) {
        resultMessage += `\n\nErrors: ${result.errors.join(', ')}`;
      }

      alert(resultMessage);

      // Reload jobs to update UI
      await loadJobs();
    } catch (error) {
      alert('Error applying to job: ' + error.message);
    } finally {
      applyingTo = null;
    }
  }
</script>

<div class="list-container">
  <div class="list-header">
    <h2>Available Jobs</h2>
    <button on:click={loadJobs} disabled={loading}>
      {loading ? 'Loading...' : 'Refresh'}
    </button>
  </div>

  {#if errorMessage}
    <p class="error-message">{errorMessage}</p>
  {/if}

  {#if loading}
    <p class="loading">Loading jobs...</p>
  {:else if jobs.length === 0}
    <p class="empty">No jobs scraped yet. Go to the Search tab to scrape jobs from Indeed.</p>
  {:else}
    <table>
      <thead>
        <tr>
          <th>Title</th>
          <th>Company</th>
          <th>Location</th>
          <th>Scraped</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        {#each jobs as job}
          <tr>
            <td>
              <a href={job.url} target="_blank" rel="noopener noreferrer">
                {job.title}
              </a>
            </td>
            <td>{job.company}</td>
            <td>{job.location || 'N/A'}</td>
            <td>{new Date(job.scraped_at).toLocaleDateString()}</td>
            <td>
              <button
                on:click={() => handleApply(job.id, job.title)}
                disabled={applyingTo === job.id}
                class="apply-btn"
              >
                {applyingTo === job.id ? 'Applying...' : 'Apply'}
              </button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  table a {
    color: #2563eb;
    text-decoration: none;
  }

  table a:hover {
    text-decoration: underline;
  }

  .apply-btn {
    padding: 0.5rem 1rem;
    font-size: 0.85rem;
    background-color: #059669;
  }

  .apply-btn:hover:not(:disabled) {
    background-color: #047857;
  }

  .apply-btn:disabled {
    background-color: #9ca3af;
  }
</style>
