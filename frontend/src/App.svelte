<script>
  import ProfileForm from './components/ProfileForm.svelte';
  import SearchConfig from './components/SearchConfig.svelte';
  import JobsList from './components/JobsList.svelte';
  import ApplicationsList from './components/ApplicationsList.svelte';
  import { hasUserId } from './lib/store';

  let currentTab = hasUserId() ? 'search' : 'profile';
</script>

<main>
  <header>
    <h1>Job Application Tool</h1>
    <nav>
      <button class:active={currentTab === 'profile'} on:click={() => currentTab = 'profile'}>
        Profile
      </button>
      <button class:active={currentTab === 'search'} on:click={() => currentTab = 'search'}>
        Search
      </button>
      <button class:active={currentTab === 'jobs'} on:click={() => currentTab = 'jobs'}>
        Jobs
      </button>
      <button class:active={currentTab === 'applications'} on:click={() => currentTab = 'applications'}>
        Applications
      </button>
    </nav>
  </header>

  <div class="container">
    {#if currentTab === 'profile'}
      <ProfileForm />
    {:else if currentTab === 'search'}
      <SearchConfig />
    {:else if currentTab === 'jobs'}
      <JobsList />
    {:else if currentTab === 'applications'}
      <ApplicationsList />
    {/if}
  </div>
</main>
