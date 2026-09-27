package requests

import "github.com/google/uuid"

type UpdateUser struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}

type SelectCosmeticOption struct {
	OptionID uuid.UUID `json:"optionId"`
}

type ChangeProfileUsername struct {
	Username string `json:"username"`
	// Unlicensed moves a licensed profile to an unlicensed nickname: its owner does not own the Minecraft account.
	Unlicensed bool `json:"unlicensed"`
}

type SetProfilePassword struct {
	SeasonID uuid.UUID `json:"seasonId"`
	Password string    `json:"password"`
}

type ConfirmProfileVerification struct {
	Code string `json:"code"`
}

// VerificationLogin is sent by the proxy plugin for every player that logs in.
type VerificationLogin struct {
	UUID       uuid.UUID `json:"uuid"`
	Username   string    `json:"username"`
	OnlineMode bool      `json:"onlineMode"`
}
