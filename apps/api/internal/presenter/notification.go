package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentNotification(notification *domain.Notification) *responses.Notification {
	return &responses.Notification{
		ID:        notification.ID,
		Type:      string(notification.Type),
		Payload:   notification.Payload,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
	}
}

func PresentNotificationList(notifications []*domain.Notification, unreadCount int64) *responses.NotificationList {
	content := make([]*responses.Notification, len(notifications))
	for i, notification := range notifications {
		content[i] = PresentNotification(notification)
	}
	return &responses.NotificationList{Content: content, UnreadCount: unreadCount}
}
