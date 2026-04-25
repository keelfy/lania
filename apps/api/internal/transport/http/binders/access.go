package binders

import (
	"net/http"
	"strings"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

func BindObtainProfileAccesses(r *http.Request) (*commands.ObtainAccessByUsernamesCommand, error) {
	seasonID, err := BindPathVariableAsUUID(r, "seasonId")
	if err != nil {
		return nil, err
	}

	usernames, err := BindMandatoryQueryParamAsString(r, "username")
	if err != nil {
		return nil, err
	}

	authUserID, err := utils.GetUserIDFromCtx(r.Context())
	if err != nil {
		return nil, err
	}

	return &commands.ObtainAccessByUsernamesCommand{
		SeasonID:    seasonID,
		Source:      domain.AccessSourceFree,
		Usernames:   strings.Split(usernames, ","),
		OwnerUserID: authUserID,
	}, nil
}

func BindCheckUsername(r *http.Request) (*commands.CheckUsernamesCommand, error) {
	usernames, err := BindPathVariable(r, "username")
	if err != nil {
		return nil, err
	}

	seasonID, err := BindOptionalQueryParamAsUUID(r, "seasonId")
	if err != nil {
		return nil, err
	}

	return &commands.CheckUsernamesCommand{
		SeasonID:  seasonID,
		Usernames: strings.Split(usernames, ","),
	}, nil
}
