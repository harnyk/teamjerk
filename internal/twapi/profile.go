package twapi

type ProfileResponse struct {
	Person ProfilePerson `json:"person"`
	Status string        `json:"status"`
}

type ProfilePerson struct {
	AvatarURL      string `json:"avatar-url"`
	EmailAddress   string `json:"email-address"`
	UserName       string `json:"user-name"`
	ID             string `json:"id"`
	CompanyName    string `json:"company-name"`
	InstallationID string `json:"installationId"`
	CompanyID      string `json:"companyId"`
	LastName       string `json:"last-name"`
	FirstName      string `json:"first-name"`
	LengthOfDay    string `json:"lengthOfDay"`
}
