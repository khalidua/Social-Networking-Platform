package domain

type AuthUser struct {
	ID              string `json:"id"`
	Provider        string `json:"provider"`
	ProviderSubject string `json:"provider_subject"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	ProfilePicURL   string `json:"profile_picture"`
}
