package domain

import (
	"fmt"
	"strings"
)

// Verifier validates whether a user is eligible to join an organisation.
type Verifier interface {
	Verify(email string, config map[string]interface{}) error
}

type EmailDomainVerifier struct{}

// Verify accepts config {"domain": "uni.lk"} or {"domains": ["uni.lk", "gmail.com"]}.
func (v *EmailDomainVerifier) Verify(email string, config map[string]interface{}) error {
	var allowed []string
	if d, ok := config["domain"].(string); ok {
		allowed = append(allowed, d)
	}
	if ds, ok := config["domains"].([]interface{}); ok {
		for _, d := range ds {
			if s, ok := d.(string); ok {
				allowed = append(allowed, s)
			}
		}
	}
	if len(allowed) == 0 {
		return fmt.Errorf("email_domain verifier requires 'domain' or 'domains' in config")
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format")
	}
	for _, d := range allowed {
		if strings.EqualFold(parts[1], d) {
			return nil
		}
	}
	return fmt.Errorf("email domain %s is not allowed for this organisation", parts[1])
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
