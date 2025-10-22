<script>
  import { onMount } from 'svelte';
  import { signup, login, getMe } from '../lib/api';
  import { setAuthToken, getAuthToken, clearAuth, setUser, getUser } from '../lib/store';

  export let onAuthSuccess;

  let mode = 'login'; // 'login' or 'signup'
  let email = '';
  let password = '';
  let fullName = '';
  let message = '';
  let loading = false;
  let isAuthenticated = false;
  let currentUser = null;

  onMount(async () => {
    const token = getAuthToken();
    if (token) {
      try {
        const user = await getMe();
        currentUser = user;
        isAuthenticated = true;
        onAuthSuccess(user);
      } catch (error) {
        clearAuth();
      }
    }
  });

  async function handleSubmit() {
    message = '';

    if (!email || !password) {
      message = 'Email and password are required';
      return;
    }

    if (mode === 'signup' && !fullName) {
      message = 'Full name is required for signup';
      return;
    }

    loading = true;

    try {
      let response;
      if (mode === 'login') {
        response = await login(email, password);
      } else {
        response = await signup(fullName, email, password);
      }

      setAuthToken(response.token);
      setUser({
        id: response.user_id,
        email: response.email,
        name: response.name
      });

      message = mode === 'login' ? 'Login successful!' : 'Account created successfully!';

      setTimeout(() => {
        isAuthenticated = true;
        currentUser = getUser();
        onAuthSuccess(currentUser);
      }, 500);
    } catch (error) {
      message = 'Error: ' + error.message;
    } finally {
      loading = false;
    }
  }

  function handleLogout() {
    clearAuth();
    isAuthenticated = false;
    currentUser = null;
    email = '';
    password = '';
    fullName = '';
    message = '';
  }

  function switchMode() {
    mode = mode === 'login' ? 'signup' : 'login';
    message = '';
  }
</script>

{#if isAuthenticated && currentUser}
  <div class="auth-container">
    <div class="user-info">
      <h3>Welcome, {currentUser.name}!</h3>
      <p class="email">{currentUser.email}</p>
      <button class="logout-btn" on:click={handleLogout}>Logout</button>
    </div>
  </div>
{:else}
  <div class="auth-container">
    <div class="auth-box">
      <h2>{mode === 'login' ? 'Login' : 'Create Account'}</h2>

      <form on:submit|preventDefault={handleSubmit}>
        {#if mode === 'signup'}
          <div class="form-group">
            <label for="fullName">Full Name *</label>
            <input
              id="fullName"
              type="text"
              bind:value={fullName}
              placeholder="John Doe"
              disabled={loading}
            />
          </div>
        {/if}

        <div class="form-group">
          <label for="email">Email *</label>
          <input
            id="email"
            type="email"
            bind:value={email}
            placeholder="you@example.com"
            disabled={loading}
          />
        </div>

        <div class="form-group">
          <label for="password">Password *</label>
          <input
            id="password"
            type="password"
            bind:value={password}
            placeholder="At least 6 characters"
            disabled={loading}
          />
        </div>

        <button type="submit" disabled={loading}>
          {loading ? 'Processing...' : mode === 'login' ? 'Login' : 'Sign Up'}
        </button>
      </form>

      <p class="switch-mode">
        {mode === 'login' ? "Don't have an account?" : 'Already have an account?'}
        <button class="link-btn" on:click={switchMode} disabled={loading}>
          {mode === 'login' ? 'Sign up' : 'Login'}
        </button>
      </p>

      {#if message}
        <p class="message" class:error={message.startsWith('Error')} class:success={!message.startsWith('Error')}>
          {message}
        </p>
      {/if}
    </div>
  </div>
{/if}

<style>
  .auth-container {
    max-width: 400px;
    margin: 2rem auto;
    padding: 0 1rem;
  }

  .auth-box {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  }

  h2 {
    margin-top: 0;
    margin-bottom: 1.5rem;
    color: #1f2937;
    text-align: center;
  }

  .switch-mode {
    text-align: center;
    margin-top: 1.5rem;
    color: #6b7280;
    font-size: 0.9rem;
  }

  .link-btn {
    background: none;
    border: none;
    color: #3b82f6;
    cursor: pointer;
    text-decoration: underline;
    padding: 0;
    margin-left: 0.5rem;
    font-size: 0.9rem;
  }

  .link-btn:hover {
    color: #2563eb;
  }

  .link-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .message.success {
    color: #10b981;
    background-color: #d1fae5;
    border-color: #10b981;
  }

  .user-info {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    text-align: center;
  }

  .user-info h3 {
    margin-top: 0;
    color: #1f2937;
  }

  .email {
    color: #6b7280;
    margin-bottom: 1.5rem;
  }

  .logout-btn {
    background-color: #ef4444;
    color: white;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 6px;
    font-size: 1rem;
    cursor: pointer;
  }

  .logout-btn:hover {
    background-color: #dc2626;
  }
</style>
