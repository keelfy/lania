package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

// seasonIDFromRequest returns the season the seasonId query parameter names,
// and the primary season when the request names none. It fails with a not found error for an unknown season.
func seasonIDFromRequest(r *http.Request, seasons services.SeasonService) (uuid.UUID, error) {
	ctx := r.Context()

	requested, err := binders.BindOptionalQueryParamAsUUID(r, "seasonId")
	if err != nil {
		return uuid.Nil, err
	}
	if requested == nil {
		return seasons.GetPrimarySeasonID(ctx)
	}
	if _, err := seasons.GetSeasonByID(ctx, *requested); err != nil {
		return uuid.Nil, err
	}
	return *requested, nil
}

// activeSeasonIDFromRequest is seasonIDFromRequest for a request that changes what a player shows in the season.
// The server of an ended season is gone, so its selection stays as the player left it.
func activeSeasonIDFromRequest(r *http.Request, seasons services.SeasonService) (uuid.UUID, error) {
	seasonID, err := seasonIDFromRequest(r, seasons)
	if err != nil {
		return uuid.Nil, err
	}
	season, err := seasons.GetSeasonByID(r.Context(), seasonID)
	if err != nil {
		return uuid.Nil, err
	}
	if !season.IsActive {
		return uuid.Nil, utils.NewBadRequestError("cosmetics can be changed in an active season only", nil)
	}
	return seasonID, nil
}
