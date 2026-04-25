package requests

import "github.com/google/uuid"

type UpdateUser struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}

type SelectCosmeticOption struct {
	OptionID uuid.UUID `json:"optionId"`
}
