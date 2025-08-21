package user

import (
	"context"
	"time"
)

type UserService interface {
	GetUserByID(context.Context, string) (User, error)
	InsertUser(context.Context, User) error
	UpdateUser(context.Context, User) error
	DeleteUser(context.Context, string) error
	InsertAdsHistory(context.Context, string) error
	GetAdsHistory(context.Context, string) (int64, error)
	GetNotifications(context.Context, string) ([]Notification, error)
	UpdateNotificationStatus(context.Context, string, int32) error
	MarkAllAllNotificationsRead(context.Context, string, time.Time) error
}
