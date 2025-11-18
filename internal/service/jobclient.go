package service

import (
	"time"

	"pxbox/internal/jobs"

	"github.com/hibiken/asynq"
)

// JobClient interface for scheduling background jobs
type JobClient interface {
	ScheduleDeadlineNotification(requestID string, deadlineAt time.Time) error
	ScheduleDeadlineExpiry(requestID string, deadlineAt time.Time) error
	ScheduleAutoCancel(requestID string, gracePeriod time.Duration) error
	ScheduleAttentionNotification(requestID string, attentionAt time.Time) error
	ScheduleReminder(reminderID string, remindAt time.Time) error
}

// AsynqJobClient implements JobClient using asynq
type AsynqJobClient struct {
	client *asynq.Client
}

func NewAsynqJobClient(client *asynq.Client) *AsynqJobClient {
	return &AsynqJobClient{client: client}
}

func (c *AsynqJobClient) ScheduleDeadlineNotification(requestID string, deadlineAt time.Time) error {
	return jobs.ScheduleDeadlineNotification(c.client, requestID, deadlineAt)
}

func (c *AsynqJobClient) ScheduleDeadlineExpiry(requestID string, deadlineAt time.Time) error {
	return jobs.ScheduleDeadlineExpiry(c.client, requestID, deadlineAt)
}

func (c *AsynqJobClient) ScheduleAutoCancel(requestID string, gracePeriod time.Duration) error {
	return jobs.ScheduleAutoCancel(c.client, requestID, gracePeriod)
}

func (c *AsynqJobClient) ScheduleAttentionNotification(requestID string, attentionAt time.Time) error {
	return jobs.ScheduleAttentionNotification(c.client, requestID, attentionAt)
}

func (c *AsynqJobClient) ScheduleReminder(reminderID string, remindAt time.Time) error {
	return jobs.ScheduleReminder(c.client, reminderID, remindAt)
}

