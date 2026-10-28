// Mirrors the identity service's rules so the form can flag mistakes before saving.

const SLUG = /^[a-z0-9]+(-[a-z0-9]+)*$/;
const DOMAIN = /^(\*\.)?([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$/;

export function normalizeDomain(raw: string): string {
  return raw.trim().toLowerCase().replace(/^@/, '');
}

export function domainError(domain: string): string | null {
  return DOMAIN.test(domain) ? null : `"${domain}" is not a valid email domain`;
}

export function slugError(slug: string): string | null {
  if (slug.length < 2 || slug.length > 40 || !SLUG.test(slug)) {
    return 'Use 2-40 lowercase letters, digits and single hyphens, e.g. ruh-eng';
  }
  return null;
}

/** "Faculty of Engineering, Ruhuna" -> "faculty-of-engineering-ruhuna" */
export function suggestSlug(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 40)
    .replace(/-+$/, '');
}

/** An example address for the preview line: "*.ruh.ac.lk" -> "name@eng.ruh.ac.lk". */
export function exampleEmail(domain: string): string {
  return 'name@' + (domain.startsWith('*.') ? 'dept.' + domain.slice(2) : domain);
}
