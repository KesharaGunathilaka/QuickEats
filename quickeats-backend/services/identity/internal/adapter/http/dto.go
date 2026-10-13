package http

type RegisterRequest struct {
	OrganisationSlug string `json:"organisationSlug"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	DisplayName      string `json:"displayName"`
	InviteCode       string `json:"inviteCode,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type CreateOrgRequest struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	VerifyMethod string                 `json:"verifyMethod"`
	VerifyConfig map[string]interface{} `json:"verifyConfig"`
}
