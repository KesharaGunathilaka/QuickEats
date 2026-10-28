package domain

import (
	"fmt"
	"net/mail"
	"strings"
)

// Verifier validates whether a user is eligible to join an organisation.
type Verifier interface {
	Verify(email string, config map[string]interface{}) error
}

type EmailDomainVerifier struct{}

// Verify accepts config {"domain": "uni.lk"} or {"domains": ["uni.lk", "*.ruh.ac.lk"]}.
func (v *EmailDomainVerifier) Verify(email string, config map[string]interface{}) error {
	allowed := (&Organisation{VerifyConfig: config}).EmailDomains()
	if len(allowed) == 0 {
		return fmt.Errorf("email_domain verifier requires 'domain' or 'domains' in config")
	}
	domain, err := EmailDomain(email)
	if err != nil {
		return err
	}
	for _, d := range allowed {
		if DomainMatches(domain, d) {
			return nil
		}
	}
	return fmt.Errorf("email domain %s is not allowed for this organisation", domain)
}

// DomainMatches reports whether an email domain satisfies an allowed pattern.
// "*.ruh.ac.lk" matches any subdomain of ruh.ac.lk, but not ruh.ac.lk itself.
func DomainMatches(emailDomain, pattern string) bool {
	emailDomain = strings.ToLower(emailDomain)
	pattern = strings.ToLower(pattern)
	if suffix, ok := strings.CutPrefix(pattern, "*."); ok {
		return strings.HasSuffix(emailDomain, "."+suffix)
	}
	return emailDomain == pattern
}

// EmailDomain validates an email address and returns its lower-cased domain.
func EmailDomain(email string) (string, error) {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != strings.TrimSpace(email) {
		return "", fmt.Errorf("invalid email format")
	}
	at := strings.LastIndex(addr.Address, "@")
	domain := strings.ToLower(addr.Address[at+1:])
	if !strings.Contains(domain, ".") {
		return "", fmt.Errorf("invalid email format")
	}
	return domain, nil
}

type InviteCodeVerifier struct{}

func (v *InviteCodeVerifier) Verify(email string, config map[string]interface{}) error {
	// Invite code is validated separately at registration time, not via email.
	// This verifier always passes — the actual code check is in the registration use case.
	return nil
}

// ResolveVerifier returns the appropriate verifier based on the method.
func ResolveVerifier(method VerifyMethod) (Verifier, error) {
	switch method {
	case VerifyEmailDomain:
		return &EmailDomainVerifier{}, nil
	case VerifyInviteCode:
		return &InviteCodeVerifier{}, nil
	default:
		return nil, fmt.Errorf("unknown verify method: %s", method)
	}
}
