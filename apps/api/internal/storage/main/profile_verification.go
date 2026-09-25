package sql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// Expiry dates are computed by the database, so they use the same clock as the checks against them.
const upsertProfileVerification = `
INSERT INTO profile_verifications (profile_id, code, attempts, expires_at, created_at)
VALUES (?, NULL, 0, DATE_ADD(NOW(), INTERVAL ? SECOND), NOW())
ON DUPLICATE KEY UPDATE
	code = NULL,
	attempts = 0,
	expires_at = VALUES(expires_at),
	created_at = NOW()
`

func (q *queries) UpsertProfileVerification(ctx context.Context, profileID uuid.UUID, ttl time.Duration) error {
	_, err := q.x.ExecContext(ctx, upsertProfileVerification, profileID, int64(ttl.Seconds()))
	return err
}

const findOpenProfileVerification = `
SELECT profile_id, code, attempts, expires_at
FROM profile_verifications
WHERE profile_id = ? AND expires_at > NOW()
`

func (q *queries) FindOpenProfileVerification(ctx context.Context, profileID uuid.UUID) (*domain.ProfileVerification, error) {
	var verification domain.ProfileVerification
	err := q.x.QueryRowContext(ctx, findOpenProfileVerification, profileID).Scan(
		&verification.ProfileID,
		&verification.Code,
		&verification.Attempts,
		&verification.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

const setProfileVerificationCode = `
UPDATE profile_verifications
SET code = ?, expires_at = DATE_ADD(NOW(), INTERVAL ? SECOND)
WHERE profile_id = ?
`

func (q *queries) SetProfileVerificationCode(ctx context.Context, profileID uuid.UUID, code string, ttl time.Duration) error {
	_, err := q.x.ExecContext(ctx, setProfileVerificationCode, code, int64(ttl.Seconds()), profileID)
	return err
}

const incrementProfileVerificationAttempts = `
UPDATE profile_verifications SET attempts = attempts + 1 WHERE profile_id = ?
`

func (q *queries) IncrementProfileVerificationAttempts(ctx context.Context, profileID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, incrementProfileVerificationAttempts, profileID)
	return err
}

const deleteProfileVerification = `
DELETE FROM profile_verifications WHERE profile_id = ?
`

func (q *queries) DeleteProfileVerification(ctx context.Context, profileID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, deleteProfileVerification, profileID)
	return err
}

const setProfileVerified = `
UPDATE profiles SET verified_mc_uuid = ?, verified_at = NOW(), updated_at = NOW() WHERE id = ?
`

func (q *queries) SetProfileVerified(ctx context.Context, profileID, mcUUID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, setProfileVerified, mcUUID, profileID)
	return err
}
