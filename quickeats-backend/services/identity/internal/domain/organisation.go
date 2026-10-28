package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type VerifyMethod string

const (
	VerifyEmailDomain VerifyMethod = "email_domain"
	VerifyInviteCode  VerifyMethod = "invite_code"
)

type Organisation struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	VerifyMethod VerifyMethod           `json:"verifyMethod"`
	VerifyConfig map[string]interface{} `json:"verifyConfig"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// NewEmailDomainOrganisation builds an organisation whose members sign up with an
// email on one of the given domains (the admin dashboard's only kind).
func NewEmailDomainOrganisation(name, slug string, domains []string) *Organisation {
	now := time.Now().UTC()
	return &Organisation{
		ID:           uuid.New().String(),
		Name:         name,
		Slug:         slug,
		VerifyMethod: VerifyEmailDomain,
		VerifyConfig: map[string]interface{}{"domains": domains},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// EmailDomains lists the email domains accepted by an email_domain organisation.
func (o *Organisation) EmailDomains() []string {
	var domains []string
	if d, ok := o.VerifyConfig["domain"].(string); ok && d != "" {
		domains = append(domains, d)
	}
	switch ds := o.VerifyConfig["domains"].(type) {
	case []interface{}:
		for _, d := range ds {
			if s, ok := d.(string); ok && s != "" {
				domains = append(domains, s)
			}
		}
	case []string:
		domains = append(domains, ds...)
	}
	return domains
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// NormalizeSlug lower-cases an organisation code and checks its format (e.g. "ruh-eng").
func NormalizeSlug(slug string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(slug))
	if len(s) < 2 || len(s) > 40 || !slugPattern.MatchString(s) {
		return "", fmt.Errorf("organisation code must be 2-40 lowercase letters, digits or single hyphens (e.g. ruh-eng)")
	}
	return s, nil
}

var domainPattern = regexp.MustCompile(`^(\*\.)?([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$`)

// NormalizeDomains cleans the admin's email domain list: "@Eng.Ruh.ac.lk" -> "eng.ruh.ac.lk".
// A leading "*." also accepts every subdomain ("*.ruh.ac.lk" matches "eng.ruh.ac.lk").
func NormalizeDomains(domains []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(d))
		d = strings.TrimPrefix(d, "@")
		if d == "" || seen[d] {
			continue
		}
		if !domainPattern.MatchString(d) {
			return nil, fmt.Errorf("%q is not a valid email domain", d)
		}
		seen[d] = true
		out = append(out, d)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one email domain is required")
	}
	return out, nil
}
