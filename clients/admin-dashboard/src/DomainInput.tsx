import { KeyboardEvent, useState } from 'react';
import { domainError, normalizeDomain } from './validation';

interface Props {
  value: string[];
  onChange: (domains: string[]) => void;
}

/** Chip input for email domains: type a domain, press Enter or comma to add it. */
export default function DomainInput({ value, onChange }: Props) {
  const [draft, setDraft] = useState('');
  const [error, setError] = useState<string | null>(null);

  function commit(): boolean {
    const parts = draft.split(/[,\s]+/).map(normalizeDomain).filter(Boolean);
    if (parts.length === 0) return true;
    const bad = parts.map(domainError).find(Boolean);
    if (bad) {
      setError(bad);
      return false;
    }
    onChange([...value, ...parts.filter((p) => !value.includes(p))]);
    setDraft('');
    setError(null);
    return true;
  }

  function onKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      commit();
    } else if (e.key === 'Backspace' && draft === '' && value.length > 0) {
      onChange(value.slice(0, -1));
    }
  }

  return (
    <div>
      <div className={`chips-input${error ? ' invalid' : ''}`}>
        {value.map((d) => (
          <span className="chip" key={d}>
            @{d}
            <button type="button" aria-label={`Remove ${d}`} onClick={() => onChange(value.filter((x) => x !== d))}>
              ×
            </button>
          </span>
        ))}
        <input
          value={draft}
          placeholder={value.length ? 'Add another…' : 'eng.ruh.ac.lk'}
          onChange={(e) => {
            setDraft(e.target.value);
            setError(null);
          }}
          onKeyDown={onKeyDown}
          onBlur={commit}
          aria-label="Email domain"
        />
      </div>
      {error ? (
        <p className="error small">{error}</p>
      ) : (
        <p className="hint">Press Enter to add. Use *.ruh.ac.lk to accept every subdomain.</p>
      )}
    </div>
  );
}
