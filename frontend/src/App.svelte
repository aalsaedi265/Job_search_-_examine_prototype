<script>
  import Auth from './components/Auth.svelte';
  import ProfileForm from './components/ProfileForm.svelte';
  import ProfileDisplay from './components/ProfileDisplay.svelte';
  import SearchConfig from './components/SearchConfig.svelte';
  import JobsList from './components/JobsList.svelte';
  import ApplicationsList from './components/ApplicationsList.svelte';
  import { isAuthenticated, clearAuth } from './lib/store';

  let authenticated = isAuthenticated();
  let currentTab = 'profile';
  let profileMode = 'view'; // 'view' or 'edit'

  function handleAuthSuccess(user) {
    authenticated = true;
    // Start at profile tab - ProfileDisplay will auto-switch to edit if no profile exists
    currentTab = 'profile';
    profileMode = 'view';
  }

  function handleLogout() {
    clearAuth();
    authenticated = false;
    currentTab = 'profile';
    profileMode = 'view';
    window.location.reload();
  }

  function handleProfileSaved() {
    profileMode = 'view';
  }

  function handleProfileEdit() {
    profileMode = 'edit';
  }

  function handleProfileDeleted() {
    handleLogout();
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
      {#if profileMode === 'view'}
        {#key profileMode}
          <ProfileDisplay onEdit={handleProfileEdit} onLogout={handleProfileDeleted} />
        {/key}
      {:else}
        <ProfileForm onSaved={handleProfileSaved} />
      {/if}
    {:else if currentTab === 'search'}
      <SearchConfig />
    {:else if currentTab === 'jobs'}
      <JobsList />
    {:else if currentTab === 'applications'}
      <ApplicationsList />
    {/if}
  </div>
</main>
