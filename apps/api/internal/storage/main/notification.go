package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const insertNotification = `
INSERT INTO notifications (
	user_id,
	type,
	payload,
	created_at
) VALUES (
	?,
	?,
	?,
	now()
)
`

type InsertNotificationParams struct {
	UserID  uuid.UUID
	Type    domain.NotificationType
	Payload json.RawMessage
}

func (q *queries) InsertNotification(ctx context.Context, arg InsertNotificationParams) error {
	_, err := q.x.ExecContext(ctx, insertNotification,
		arg.UserID,
		string(arg.Type),
		[]byte(arg.Payload),
	)
	return err
}

const findNotificationsByUserID = `
SELECT
	id,
	user_id,
	type,
	payload,
	read_at,
	created_at
FROM notifications
WHERE user_id = ?
ORDER BY created_at DESC, id
LIMIT ?
`

func (q *queries) FindNotificationsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Notification, error) {
	rows, err := q.x.QueryContext(ctx, findNotificationsByUserID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		var notification domain.Notification
		var notificationType string
		var payload json.RawMessage
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notificationType,
			&payload,
			&notification.ReadAt,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notification.Type = domain.NotificationType(notificationType)
		notification.Payload = payload
		notifications = append(notifications, &notification)
	}
	return notifications, rows.Err()
}

const countUnreadNotificationsByUserID = `
SELECT COUNT(*)
FROM notifications
WHERE user_id = ? AND read_at IS NULL
`

func (q *queries) CountUnreadNotificationsByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := q.x.QueryRowContext(ctx, countUnreadNotificationsByUserID, userID).Scan(&count)
	return count, err
}

const markNotificationsReadByUserID = `
UPDATE notifications
SET read_at = now()
WHERE user_id = ? AND read_at IS NULL
`

const markNotificationsReadByUserIDAndIDs = markNotificationsReadByUserID + `
AND id IN ('%s')
`

// MarkNotificationsRead stamps the unread notifications of the user. Empty ids marks all of them.
// A notification of another user is never touched.
func (q *queries) MarkNotificationsRead(ctx context.Context, userID uuid.UUID, ids uuid.UUIDs) error {
	if len(ids) == 0 {
		_, err := q.x.ExecContext(ctx, markNotificationsReadByUserID, userID)
		return err
	}

	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = id.String()
	}
	query := fmt.Sprintf(markNotificationsReadByUserIDAndIDs, strings.Join(idsStr, "','"))
	_, err := q.x.ExecContext(ctx, query, userID)
	return err
}

const deleteNotificationsByUserID = `
DELETE FROM notifications
WHERE user_id = ?
`

func (q *queries) DeleteNotificationsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, deleteNotificationsByUserID, userID)
	return err
}
