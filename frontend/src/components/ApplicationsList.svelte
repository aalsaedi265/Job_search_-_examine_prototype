<script>
  import { onMount } from 'svelte';
  import { getApplications } from '../lib/api';
  import { getUserId } from '../lib/store';

  let applications = [];
  let loading = false;
  let errorMessage = '';

  onMount(async () => {
    await loadApplications();
  });

  async function loadApplications() {
    const userId = getUserId();
    if (!userId) {
      errorMessage = 'Please create a profile first';
      return;
    }

    loading = true;
    errorMessage = '';
    try {
      applications = await getApplications(userId);
    } catch (error) {
      errorMessage = 'Error loading applications: ' + error.message;
      applications = [];
    } finally {
      loading = false;
    }
  }
</script>

<div class="list-container">
  <div class="list-header">
    <h2>Your Applications</h2>
    <button on:click={loadApplications} disabled={loading}>
      {loading ? 'Loading...' : 'Refresh'}
    </button>
  </div>

  {#if errorMessage}
    <p class="error-message">{errorMessage}</p>
  {/if}

  {#if loading}
    <p class="loading">Loading applications...</p>
  {:else if applications.length === 0}
    <p class="empty">No applications yet. Apply to jobs from the Jobs tab.</p>
  {:else}
    <table>
      <thead>
        <tr>
          <th>Job Title</th>
          <th>Company</th>
          <th>Status</th>
          <th>Fields Filled</th>
          <th>Applied Date</th>
        </tr>
      </thead>
      <tbody>
        {#each applications as app}
          <tr>
            <td>
              {#if app.job_url}
                <a href={app.job_url} target="_blank" rel="noopener noreferrer">
                  {app.job_title}
                </a>
              {:else}
                {app.job_title}
              {/if}
            </td>
            <td>{app.company}</td>
            <td>
              <span class="status status-{app.status}">
                {app.status}
              </span>
            </td>
            <td>
              {#if app.fields_filled && app.fields_filled.length > 0}
                <span class="fields">{app.fields_filled.join(', ')}</span>
              {:else}
                <span class="no-data">-</span>
              {/if}
            </td>
            <td>{new Date(app.applied_at).toLocaleString()}</td>
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

  .status {
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .status-submitted {
    background-color: #d1fae5;
    color: #065f46;
  }

  .status-failed {
    background-color: #fee2e2;
    color: #991b1b;
  }

  .fields {
    font-size: 0.85rem;
    color: #6b7280;
  }

  .no-data {
    color: #9ca3af;
  }
</style>
