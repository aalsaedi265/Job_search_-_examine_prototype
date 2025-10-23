<script>
  import Auth from './components/Auth.svelte';
  import ProfileForm from './components/ProfileForm.svelte';
  import SearchConfig from './components/SearchConfig.svelte';
  import JobsList from './components/JobsList.svelte';
  import ApplicationsList from './components/ApplicationsList.svelte';
  import { isAuthenticated, clearAuth } from './lib/store';

  let authenticated = isAuthenticated();
  let currentTab = 'profile';

  function handleAuthSuccess(user) {
    authenticated = true;
    currentTab = 'search';
  }

  function handleLogout() {
    clearAuth();
    authenticated = false;
    currentTab = 'profile';
    window.location.reload();
  }
</script>

<main>
  <header>
    <div class="header-content">
      <h1>Job Application Tool</h1>
      {#if authenticated}
        <button class="logout-btn" on:click={handleLogout}>
          Logout
        </button>
      {/if}
    </div>
    {#if authenticated}
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
    {/if}
  </header>

  <div class="container">
    {#if !authenticated}
      <Auth onAuthSuccess={handleAuthSuccess} />
    {:else if currentTab === 'profile'}
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
