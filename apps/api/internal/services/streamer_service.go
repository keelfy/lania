package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// StreamerService gives out the streamer role. An owner applies for it, an admin approves or rejects the
// application, or grants and revokes the role directly. A change of the role stamps the role of the profile,
// so role sync puts the player into the streamer group of the season servers.
type StreamerService interface {
	// Apply sends an application of the owner for the streamer role of the profile. It fails when the profile is
	// already a streamer, has a pending application, or was rejected less than domain.StreamerReapplyCooldown ago.
	Apply(ctx context.Context, userID, profileID uuid.UUID, channels []domain.StreamerChannel, about string) (*domain.StreamerApplication, error)
	// GetOwnedStreamer returns the streamer card of the profile, nil when it is not a streamer, and the latest
	// application, nil when the owner never applied.
	GetOwnedStreamer(ctx context.Context, userID, profileID uuid.UUID) (*domain.Streamer, *domain.StreamerApplication, error)
	// UpdateOwnedStreamer replaces the channels and description of the owner's streamer card.
	UpdateOwnedStreamer(ctx context.Context, userID, profileID uuid.UUID, channels []domain.StreamerChannel, description string) (*domain.Streamer, error)
	// GetStreamers returns every streamer with its profile, the earliest first.
	GetStreamers(ctx context.Context) ([]*domain.Streamer, error)

	// GetApplications returns a page of applications with the status and how many there are.
	GetApplications(ctx context.Context, status domain.StreamerApplicationStatus, pagination *domain.Pagination) ([]*domain.StreamerApplication, int64, error)
	// ApproveApplication makes the profile of a pending application a streamer with the channels of the application.
	ApproveApplication(ctx context.Context, applicationID, adminID uuid.UUID) error
	// RejectApplication turns a pending application down with a reason the owner sees.
	RejectApplication(ctx context.Context, applicationID, adminID uuid.UUID, reason string) error
	// GetStreamer returns the streamer card of the profile, nil when it is not a streamer.
	GetStreamer(ctx context.Context, profileID uuid.UUID) (*domain.Streamer, error)
	// SaveStreamer makes the profile a streamer without an application, or edits its card.
	SaveStreamer(ctx context.Context, adminID, profileID uuid.UUID, channels []domain.StreamerChannel, description string) (*domain.Streamer, error)
	// RevokeStreamer takes the streamer role away and tells the owner.
	RevokeStreamer(ctx context.Context, adminID, profileID uuid.UUID) error
}

type streamerService struct {
	storage             storage.MainStorage
	notificationService NotificationService

	now func() time.Time
}

func NewStreamerService(storage storage.MainStorage, notificationService NotificationService) StreamerService {
	return &streamerService{storage: storage, notificationService: notificationService, now: time.Now}
}

// normalizeStreamerCard trims the input and checks it. An empty description is none.
func normalizeStreamerCard(channels []domain.StreamerChannel, description string) ([]domain.StreamerChannel, *string, error) {
	normalized := make([]domain.StreamerChannel, len(channels))
	for i, channel := range channels {
		normalized[i] = domain.StreamerChannel{Platform: channel.Platform, URL: strings.TrimSpace(channel.URL)}
	}
	if err := domain.ValidateStreamerChannels(normalized); err != nil {
		return nil, nil, utils.NewBadRequestError(err.Error(), err)
	}

	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > domain.MaxStreamerDescriptionLength {
		return nil, nil, utils.NewBadRequestError("the description is too long", nil)
	}
	if description == "" {
		return normalized, nil, nil
	}
	return normalized, &description, nil
}

// ownedProfile returns the profile when the user owns it.
func (s *streamerService) ownedProfile(ctx context.Context, queries sql.Queries, userID, profileID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.findProfile(ctx, queries, profileID)
	if err != nil {
		return nil, err
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return nil, utils.NewForbiddenError("only owner can manage the streamer role of the profile", nil)
	}
	return profile, nil
}

func (s *streamerService) findProfile(ctx context.Context, queries sql.Queries, profileID uuid.UUID) (*domain.Profile, error) {
	profile, err := queries.FindProfileByID(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile", err)
	}
	return profile, nil
}

// findStreamer returns nil when the profile is not a streamer.
func findStreamer(ctx context.Context, queries sql.Queries, profileID uuid.UUID) (*domain.Streamer, error) {
	streamer, err := queries.FindStreamer(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get streamer", err)
	}
	return streamer, nil
}

// findLatestApplication returns nil when the profile never applied.
func findLatestApplication(ctx context.Context, queries sql.Queries, profileID uuid.UUID) (*domain.StreamerApplication, error) {
	application, err := queries.FindLatestStreamerApplication(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get streamer application", err)
	}
	return application, nil
}

func (s *streamerService) Apply(ctx context.Context, userID, profileID uuid.UUID, channels []domain.StreamerChannel, about string) (*domain.StreamerApplication, error) {
	channels, _, err := normalizeStreamerCard(channels, "")
	if err != nil {
		return nil, err
	}
	about = strings.TrimSpace(about)
	if about == "" {
		return nil, utils.NewBadRequestError("tell about your streams", nil)
	}
	if utf8.RuneCountInString(about) > domain.MaxStreamerAboutLength {
		return nil, utils.NewBadRequestError("the text about your streams is too long", nil)
	}

	application := &domain.StreamerApplication{
		ID:        uuid.New(),
		ProfileID: profileID,
		UserID:    userID,
		Channels:  channels,
		About:     about,
		Status:    domain.StreamerApplicationStatusPending,
		CreatedAt: s.now(),
	}
	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		// The lock keeps two applications of the profile from passing the checks at once.
		if err := queries.LockProfile(ctx, profileID); err != nil {
			return utils.NewInternalServerError("failed to lock profile", err)
		}
		if _, err := s.ownedProfile(ctx, queries, userID, profileID); err != nil {
			return err
		}

		streamer, err := findStreamer(ctx, queries, profileID)
		if err != nil {
			return err
		}
		if streamer != nil {
			return utils.NewConflictError("the profile is already a streamer", nil)
		}
		latest, err := findLatestApplication(ctx, queries, profileID)
		if err != nil {
			return err
		}
		if latest != nil && latest.Status == domain.StreamerApplicationStatusPending {
			return utils.NewConflictError("the profile already has an application under review", nil)
		}
		if latest != nil {
			if reapplyAt := latest.ReapplyAt(); reapplyAt != nil && s.now().Before(*reapplyAt) {
				return utils.NewTooManyRequestsError("you can apply again on "+reapplyAt.UTC().Format(time.DateOnly), nil)
			}
		}

		if err := queries.InsertStreamerApplication(ctx, application); err != nil {
			return utils.NewInternalServerError("failed to store streamer application", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (s *streamerService) GetOwnedStreamer(ctx context.Context, userID, profileID uuid.UUID) (*domain.Streamer, *domain.StreamerApplication, error) {
	queries := s.storage.Queries()
	if _, err := s.ownedProfile(ctx, queries, userID, profileID); err != nil {
		return nil, nil, err
	}
	streamer, err := findStreamer(ctx, queries, profileID)
	if err != nil {
		return nil, nil, err
	}
	application, err := findLatestApplication(ctx, queries, profileID)
	if err != nil {
		return nil, nil, err
	}
	return streamer, application, nil
}

func (s *streamerService) UpdateOwnedStreamer(ctx context.Context, userID, profileID uuid.UUID, channels []domain.StreamerChannel, description string) (*domain.Streamer, error) {
	channels, desc, err := normalizeStreamerCard(channels, description)
	if err != nil {
		return nil, err
	}

	queries := s.storage.Queries()
	if _, err := s.ownedProfile(ctx, queries, userID, profileID); err != nil {
		return nil, err
	}
	streamer, err := findStreamer(ctx, queries, profileID)
	if err != nil {
		return nil, err
	}
	if streamer == nil {
		return nil, utils.NewForbiddenError("the profile is not a streamer", nil)
	}

	// Only the card changes, the group on the servers stays, so the role is not stamped.
	if err := queries.SaveStreamer(ctx, sql.SaveStreamerParams{ProfileID: profileID, Channels: channels, Description: desc}); err != nil {
		return nil, utils.NewInternalServerError("failed to save streamer", err)
	}
	return findStreamer(ctx, queries, profileID)
}

func (s *streamerService) GetStreamers(ctx context.Context) ([]*domain.Streamer, error) {
	streamers, err := s.storage.Queries().FindStreamers(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get streamers", err)
	}
	return streamers, nil
}

func (s *streamerService) GetApplications(ctx context.Context, status domain.StreamerApplicationStatus, pagination *domain.Pagination) ([]*domain.StreamerApplication, int64, error) {
	if !status.Valid() {
		return nil, 0, utils.NewBadRequestError("unknown application status", nil)
	}
	applications, err := s.storage.Queries().FindStreamerApplications(ctx, status, pagination.Size, pagination.From)
	if err != nil {
		return nil, 0, utils.NewInternalServerError("failed to get streamer applications", err)
	}
	count, err := s.storage.Queries().CountStreamerApplications(ctx, status)
	if err != nil {
		return nil, 0, utils.NewInternalServerError("failed to count streamer applications", err)
	}
	return applications, count, nil
}

// reviewApplication closes a pending application inside a transaction that holds the profile. apply runs after
// the application is closed and before the transaction ends.
func (s *streamerService) reviewApplication(
	ctx context.Context,
	applicationID, adminID uuid.UUID,
	status domain.StreamerApplicationStatus,
	reason *string,
	apply func(queries sql.Queries, application *domain.StreamerApplication, profile *domain.Profile) error,
) error {
	application, err := s.storage.Queries().FindStreamerApplicationByID(ctx, applicationID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return utils.NewNotFoundError("streamer application not found", err)
	} else if err != nil {
		return utils.NewInternalServerError("failed to get streamer application", err)
	}

	return s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.LockProfile(ctx, application.ProfileID); err != nil {
			return utils.NewInternalServerError("failed to lock profile", err)
		}
		profile, err := s.findProfile(ctx, queries, application.ProfileID)
		if err != nil {
			return err
		}
		reviewed, err := queries.ReviewStreamerApplication(ctx, applicationID, status, reason, adminID)
		if err != nil {
			return utils.NewInternalServerError("failed to review streamer application", err)
		}
		if !reviewed {
			return utils.NewConflictError("the application is already reviewed", nil)
		}
		return apply(queries, application, profile)
	})
}

func (s *streamerService) ApproveApplication(ctx context.Context, applicationID, adminID uuid.UUID) error {
	return s.reviewApplication(ctx, applicationID, adminID, domain.StreamerApplicationStatusApproved, nil,
		func(queries sql.Queries, application *domain.StreamerApplication, profile *domain.Profile) error {
			current, err := findStreamer(ctx, queries, profile.ID)
			if err != nil {
				return err
			}
			// An admin may have made the profile a streamer meanwhile: its card stays.
			if current == nil {
				if err := s.grant(ctx, queries, adminID, profile.ID, application.Channels, nil); err != nil {
					return err
				}
			}
			s.notificationService.NotifyStreamer(ctx, queries, profile, domain.NotificationTypeStreamerApproved, "")
			return nil
		})
}

func (s *streamerService) RejectApplication(ctx context.Context, applicationID, adminID uuid.UUID, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return utils.NewBadRequestError("a rejection needs a reason", nil)
	}
	if utf8.RuneCountInString(reason) > domain.MaxStreamerRejectReason {
		return utils.NewBadRequestError("the reason is too long", nil)
	}
	return s.reviewApplication(ctx, applicationID, adminID, domain.StreamerApplicationStatusRejected, &reason,
		func(queries sql.Queries, _ *domain.StreamerApplication, profile *domain.Profile) error {
			s.notificationService.NotifyStreamer(ctx, queries, profile, domain.NotificationTypeStreamerRejected, reason)
			return nil
		})
}

func (s *streamerService) GetStreamer(ctx context.Context, profileID uuid.UUID) (*domain.Streamer, error) {
	return findStreamer(ctx, s.storage.Queries(), profileID)
}

func (s *streamerService) SaveStreamer(ctx context.Context, adminID, profileID uuid.UUID, channels []domain.StreamerChannel, description string) (*domain.Streamer, error) {
	channels, desc, err := normalizeStreamerCard(channels, description)
	if err != nil {
		return nil, err
	}

	var saved *domain.Streamer
	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if _, err := s.findProfile(ctx, queries, profileID); err != nil {
			return err
		}
		if err := s.grant(ctx, queries, adminID, profileID, channels, desc); err != nil {
			return err
		}
		saved, err = findStreamer(ctx, queries, profileID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

// grant stores the streamer card and stamps the role, so role sync puts the player into the streamer group.
func (s *streamerService) grant(ctx context.Context, queries sql.Queries, adminID, profileID uuid.UUID, channels []domain.StreamerChannel, description *string) error {
	err := queries.SaveStreamer(ctx, sql.SaveStreamerParams{ProfileID: profileID, Channels: channels, Description: description, CreatedBy: &adminID})
	if err != nil {
		return utils.NewInternalServerError("failed to save streamer", err)
	}
	if err := queries.TouchProfileRole(ctx, profileID); err != nil {
		return utils.NewInternalServerError("failed to mark profile role as changed", err)
	}
	return nil
}

func (s *streamerService) RevokeStreamer(ctx context.Context, adminID, profileID uuid.UUID) error {
	return s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		profile, err := s.findProfile(ctx, queries, profileID)
		if err != nil {
			return err
		}
		revoked, err := queries.DeleteStreamer(ctx, profileID)
		if err != nil {
			return utils.NewInternalServerError("failed to revoke streamer", err)
		}
		if !revoked {
			return utils.NewNotFoundError("the profile is not a streamer", nil)
		}
		if err := queries.TouchProfileRole(ctx, profileID); err != nil {
			return utils.NewInternalServerError("failed to mark profile role as changed", err)
		}
		s.notificationService.NotifyStreamer(ctx, queries, profile, domain.NotificationTypeStreamerRevoked, "")
		return nil
	})
}
