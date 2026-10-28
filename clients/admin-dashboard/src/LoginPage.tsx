import { FormEvent, useState } from 'react';
import { login } from './api';

export default function LoginPage({ onSignedIn }: { onSignedIn: () => void }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await login(email.trim(), password);
      onSignedIn();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sign-in failed');
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="login">
      <form className="card login-card" onSubmit={submit}>
        <div className="brand">
          <span className="brand-mark" aria-hidden>Q</span>
          <div>
            <h1>QuickEats Admin</h1>
            <p className="muted">Manage organisations and who can sign up</p>
          </div>
        </div>

        <label className="field">
          <span>Email</span>
          <input type="email" autoComplete="username" required value={email} onChange={(e) => setEmail(e.target.value)} />
        </label>
        <label className="field">
          <span>Password</span>
          <input type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
        </label>

        {error && <p className="error" role="alert">{error}</p>}

        <button className="btn primary block" disabled={busy}>
          {busy ? 'Signing in…' : 'Sign in'}
        </button>
      </form>
    </main>
  );
}
