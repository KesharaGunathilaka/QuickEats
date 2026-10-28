import { useCallback, useEffect, useMemo, useState } from 'react';
import { api, Organisation } from './api';
import OrgForm from './OrgForm';

export default function OrganisationsPage({ onSignOut }: { onSignOut: () => void }) {
  const [orgs, setOrgs] = useState<Organisation[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState('');
  const [editing, setEditing] = useState<Organisation | 'new' | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      setOrgs(await api.listOrganisations());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load organisations');
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!notice) return;
    const t = setTimeout(() => setNotice(null), 3500);
    return () => clearTimeout(t);
  }, [notice]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!orgs || !q) return orgs ?? [];
    return orgs.filter(
      (o) => o.name.toLowerCase().includes(q) || o.slug.includes(q) || o.emailDomains.some((d) => d.includes(q)),
    );
  }, [orgs, query]);

  const totalMembers = orgs?.reduce((n, o) => n + (o.memberCount ?? 0), 0) ?? 0;

  return (
    <div className="shell">
      <header className="topbar">
        <div className="brand">
          <span className="brand-mark" aria-hidden>Q</span>
          <strong>QuickEats Admin</strong>
        </div>
        <button className="btn ghost" onClick={onSignOut}>
          Sign out
        </button>
      </header>

      <main className="content">
        <div className="page-head">
          <div>
            <h1>Organisations</h1>
            <p className="muted">
              Customers register with an organisation code and an email on one of its domains. Merchants pick one or
              more codes for their shop.
            </p>
          </div>
          <button className="btn primary" onClick={() => setEditing('new')}>
            + New organisation
          </button>
        </div>

        <section className="stats">
          <div className="stat card">
            <span className="stat-value">{orgs?.length ?? '–'}</span>
            <span className="muted">Organisations</span>
          </div>
          <div className="stat card">
            <span className="stat-value">{orgs ? totalMembers : '–'}</span>
            <span className="muted">Registered members</span>
          </div>
        </section>

        <div className="card table-card">
          <div className="table-tools">
            <input
              type="search"
              placeholder="Search by name, code or domain"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              aria-label="Search organisations"
            />
          </div>

          {error && (
            <div className="empty">
              <p className="error">{error}</p>
              <button className="btn" onClick={load}>
                Try again
              </button>
            </div>
          )}
          {!error && orgs === null && <div className="empty muted">Loading…</div>}
          {!error && orgs !== null && filtered.length === 0 && (
            <div className="empty muted">{orgs.length === 0 ? 'No organisations yet. Create the first one.' : 'No matches.'}</div>
          )}

          {!error && filtered.length > 0 && (
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Code</th>
                  <th>Email domains</th>
                  <th className="num">Members</th>
                  <th aria-label="Actions" />
                </tr>
              </thead>
              <tbody>
                {filtered.map((o) => (
                  <tr key={o.id}>
                    <td data-label="Name">{o.name}</td>
                    <td data-label="Code">
                      <code className="code">{o.slug}</code>
                    </td>
                    <td data-label="Email domains">
                      <div className="chips">
                        {o.emailDomains.map((d) => (
                          <span className="chip" key={d}>
                            @{d}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td data-label="Members" className="num">
                      {o.memberCount ?? 0}
                    </td>
                    <td className="actions">
                      <button className="btn small" onClick={() => setEditing(o)}>
                        Edit
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>

      {editing && (
        <OrgForm
          org={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={(saved) => {
            setEditing(null);
            setNotice(editing === 'new' ? `Created ${saved.name} (${saved.slug})` : `Saved ${saved.name}`);
            load();
          }}
        />
      )}

      {notice && (
        <div className="toast" role="status">
          {notice}
        </div>
      )}
    </div>
  );
}
