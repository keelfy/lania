package requests

type OryUserTraits struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type AfterOryRegistrationWebhook struct {
	IdentityID string         `json:"identityId"`
	Traits     *OryUserTraits `json:"traits"`
}
