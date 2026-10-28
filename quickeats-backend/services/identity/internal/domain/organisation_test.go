package domain

import "testing"

func TestEmailDomainVerifier(t *testing.T) {
	config := map[string]interface{}{"domains": []interface{}{"eng.ruh.ac.lk", "*.cmb.ac.lk"}}
	v := &EmailDomainVerifier{}

	cases := []struct {
		email string
		ok    bool
	}{
		{"student@eng.ruh.ac.lk", true},
		{"Student@ENG.Ruh.AC.lk", true},       // case-insensitive
		{"student@ruh.ac.lk", false},          // parent domain is not the faculty domain
		{"student@fake-eng.ruh.ac.lk", false}, // exact match only
		{"student@sci.cmb.ac.lk", true},       // wildcard subdomain
		{"student@cmb.ac.lk", false},          // wildcard needs a subdomain
		{"student@evilcmb.ac.lk", false},      // wildcard is not a plain suffix
		{"student@gmail.com", false},
		{"not-an-email", false},
		{"two@@eng.ruh.ac.lk", false},
	}
	for _, c := range cases {
		err := v.Verify(c.email, config)
		if (err == nil) != c.ok {
			t.Errorf("Verify(%q) error = %v, want ok=%v", c.email, err, c.ok)
		}
	}
}

func TestNormalizeDomains(t *testing.T) {
	got, err := NormalizeDomains([]string{" @Eng.Ruh.AC.LK ", "eng.ruh.ac.lk", "*.cmb.ac.lk", ""})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"eng.ruh.ac.lk", "*.cmb.ac.lk"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}

	for _, bad := range [][]string{{}, {"localhost"}, {"eng ruh.ac.lk"}, {"*.lk."}, {"user@ruh.ac.lk"}} {
		if _, err := NormalizeDomains(bad); err == nil {
			t.Errorf("NormalizeDomains(%v) should fail", bad)
		}
	}
}

func TestNormalizeSlug(t *testing.T) {
	for in, want := range map[string]string{"ruh-eng": "ruh-eng", " RUH-ENG ": "ruh-eng", "demo": "demo"} {
		if got, err := NormalizeSlug(in); err != nil || got != want {
			t.Errorf("NormalizeSlug(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "a", "ruh eng", "ruh--eng", "-ruh", "ruh_eng"} {
		if _, err := NormalizeSlug(bad); err == nil {
			t.Errorf("NormalizeSlug(%q) should fail", bad)
		}
	}
}
