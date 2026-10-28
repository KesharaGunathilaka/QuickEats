import { FormEvent, useEffect, useRef, useState } from 'react';
import { api, Organisation } from './api';
import DomainInput from './DomainInput';
import { exampleEmail, slugError, suggestSlug } from './validation';

interface Props {
  org: Organisation | null; // null = create
  onSaved: (org: Organisation) => void;
  onClose: () => void;
}

export default function OrgForm({ org, onSaved, onClose }: Props) {
  const [name, setName] = useState(org?.name ?? '');
  const [slug, setSlug] = useState(org?.slug ?? '');
  const [slugTouched, setSlugTouched] = useState(!!org);
  const [domains, setDomains] = useState<string[]>(org?.emailDomains ?? []);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const dialog = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    dialog.current?.showModal();
  }, []);

  const slugProblem = slug ? slugError(slug) : null;
  const canSave = name.trim() !== '' && slug !== '' && !slugProblem && domains.length > 0 && !busy;

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (!canSave) return;
    setBusy(true);
    setError(null);
    try {
      const input = { name: name.trim(), slug, emailDomains: domains };
      const saved = org ? await api.updateOrganisation(org.id, input) : await api.createOrganisation(input);
      onSaved({ ...org, ...saved });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save');
      setBusy(false);
    }
  }

  return (
    <dialog ref={dialog} className="dialog" onClose={onClose}>
      <form onSubmit={submit}>
        <header className="dialog-head">
          <h2>{org ? 'Edit organisation' : 'New organisation'}</h2>
          <button type="button" className="icon-btn" aria-label="Close" onClick={onClose}>
            ×
          </button>
        </header>

        <label className="field">
          <span>Name</span>
          <input
            required
            autoFocus
            value={name}
            placeholder="Faculty of Engineering, University of Ruhuna"
            onChange={(e) => {
              setName(e.target.value);
              if (!slugTouched) setSlug(suggestSlug(e.target.value));
            }}
          />
        </label>

        <label className="field">
          <span>Organisation code</span>
          <input
            required
            value={slug}
            placeholder="ruh-eng"
            className={slugProblem ? 'invalid' : ''}
            onChange={(e) => {
              setSlugTouched(true);
              setSlug(e.target.value.toLowerCase().replace(/\s+/g, '-'));
            }}
          />
          {slugProblem ? (
            <p className="error small">{slugProblem}</p>
          ) : (
            <p className="hint">Users type this code when they register.</p>
          )}
        </label>

        <div className="field">
          <span>Allowed email domains</span>
          <DomainInput value={domains} onChange={setDomains} />
        </div>

        {domains.length > 0 && (
          <div className="preview">
            Customers sign up with an email like{' '}
            {domains.slice(0, 3).map((d, i) => (
              <span key={d}>
                {i > 0 && ' or '}
                <code>{exampleEmail(d)}</code>
              </span>
            ))}
            . Merchants can use any email.
          </div>
        )}

        {org && <p className="hint">Changing domains affects new sign-ups only; existing members keep their accounts.</p>}
        {error && <p className="error" role="alert">{error}</p>}

        <footer className="dialog-foot">
          <button type="button" className="btn" onClick={onClose}>
            Cancel
          </button>
          <button className="btn primary" disabled={!canSave}>
            {busy ? 'Saving…' : org ? 'Save changes' : 'Create organisation'}
          </button>
        </footer>
      </form>
    </dialog>
  );
}
