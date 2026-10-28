import { useEffect, useState } from 'react';
import { currentSession, logout, setSessionExpiredHandler } from './api';
import LoginPage from './LoginPage';
import OrganisationsPage from './OrganisationsPage';

export default function App() {
  const [signedIn, setSignedIn] = useState(() => currentSession()?.role === 'admin');

  useEffect(() => setSessionExpiredHandler(() => setSignedIn(false)), []);

  if (!signedIn) return <LoginPage onSignedIn={() => setSignedIn(true)} />;

  return (
    <OrganisationsPage
      onSignOut={() => {
        logout();
        setSignedIn(false);
      }}
    />
  );
}
