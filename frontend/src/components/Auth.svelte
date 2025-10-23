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
    max-width: 650px;
    width: 100%;
    margin: 0 auto;
    padding: 0 1rem;
  }

  .auth-box {
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(5px);
    padding: 3rem;
    border: 3px solid #fce700;
    border-left: 5px solid #ff006e;
    box-shadow:
      0 0 50px rgba(252, 231, 0, 0.4),
      inset 0 0 40px rgba(0, 0, 0, 0.6),
      inset -3px 0 15px rgba(252, 231, 0, 0.2);
    position: relative;
    clip-path: polygon(15px 0, 100% 0, 100% calc(100% - 15px), calc(100% - 15px) 100%, 0 100%, 0 15px);
  }

  .auth-box::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, #fce700, #00f0ff, #ff006e, transparent);
    background-size: 200% 100%;
    animation: borderScan 3s linear infinite;
  }

  .auth-box::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, #ff006e, #00f0ff, #fce700, transparent);
    background-size: 200% 100%;
    animation: borderScan 3s linear infinite reverse;
  }

  @keyframes borderScan {
    0% { background-position: 0% 0%; }
    100% { background-position: 200% 0%; }
  }

  h2 {
    margin-top: 0;
    margin-bottom: 2.5rem;
    font-family: 'Teko', sans-serif;
    font-weight: 700;
    font-size: 3.5rem;
    color: #fce700;
    text-align: center;
    text-transform: uppercase;
    letter-spacing: 8px;
    text-shadow:
      0 0 20px #fce700,
      0 0 40px rgba(252, 231, 0, 0.5),
      3px 3px 0 #ff006e;
    line-height: 1;
  }

  .switch-mode {
    text-align: center;
    margin-top: 2.5rem;
    color: #ffffff;
    font-size: 1.2rem;
    font-weight: 600;
    font-family: 'Saira Condensed', sans-serif;
    letter-spacing: 1px;
  }

  .link-btn {
    background: transparent;
    border: none;
    border-bottom: 3px solid #fce700;
    color: #fce700;
    cursor: pointer;
    padding: 0.3rem 0.8rem;
    margin-left: 0.8rem;
    font-size: 1.2rem;
    font-weight: 900;
    text-transform: uppercase;
    letter-spacing: 2px;
    font-family: 'Saira Condensed', sans-serif;
    transition: all 0.2s;
  }

  .link-btn:hover {
    color: #ffffff;
    border-bottom-color: #ffffff;
    text-shadow: 0 0 20px rgba(252, 231, 0, 1);
    transform: translateY(-2px);
  }

  .link-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .message.success {
    color: #00ff9f;
    background: rgba(0, 255, 159, 0.1);
    border-color: #00ff9f;
    text-shadow: 0 0 10px rgba(0, 255, 159, 0.5);
  }

  .user-info {
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(5px);
    padding: 3rem;
    border: 3px solid #fce700;
    border-left: 5px solid #00f0ff;
    box-shadow: 0 0 50px rgba(252, 231, 0, 0.4);
    text-align: center;
    clip-path: polygon(15px 0, 100% 0, 100% calc(100% - 15px), calc(100% - 15px) 100%, 0 100%, 0 15px);
  }

  .user-info h3 {
    margin-top: 0;
    font-family: 'Teko', sans-serif;
    font-size: 3rem;
    font-weight: 700;
    color: #fce700;
    text-transform: uppercase;
    letter-spacing: 6px;
    text-shadow:
      0 0 20px #fce700,
      0 0 40px rgba(252, 231, 0, 0.5),
      2px 2px 0 #00f0ff;
    line-height: 1;
  }

  .email {
    color: #ffffff;
    font-size: 1.3rem;
    font-weight: 600;
    margin-bottom: 2.5rem;
    margin-top: 1.5rem;
    letter-spacing: 2px;
    font-family: 'Saira Condensed', sans-serif;
  }

  .logout-btn {
    background: #ff003c;
    color: #ffffff;
    padding: 1rem 2.5rem;
    border: 3px solid #ff003c;
    font-family: 'Saira Condensed', sans-serif;
    font-size: 1.3rem;
    font-weight: 900;
    text-transform: uppercase;
    letter-spacing: 3px;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow:
      0 0 30px rgba(255, 0, 60, 0.6),
      0 5px 0 #990025,
      0 7px 15px rgba(0, 0, 0, 0.5);
    clip-path: polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px);
  }

  .logout-btn:hover {
    background: #ffffff;
    color: #ff003c;
    border-color: #ff003c;
    box-shadow:
      0 0 50px rgba(255, 0, 60, 1),
      0 3px 0 #990025,
      0 5px 15px rgba(0, 0, 0, 0.6);
    transform: translateY(-3px);
  }
</style>
