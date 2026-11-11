package domain

import "time"

type ObjectStatus string

const (
	StatusTracking        ObjectStatus = "TRACKING"
	StatusSuspicious      ObjectStatus = "SUSPICIOUS"
	StatusWaitingPostRoll ObjectStatus = "WAITING_POST_ROLL"
	StatusCompleted       ObjectStatus = "COMPLETED"
)

type DetectedObject struct {
	Uuid       string       `json:"uuid"`
	ClassId    int          `json:"class_id"`
	ClassName  string       `json:"class_name"`
	Confidence float32      `json:"confidence"`
	X          int          `json:"x"`
	Y          int          `json:"y"`
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	CreatedAt  time.Time    `json:"created_at"`
	LastSeenAt time.Time    `json:"last_seen_at"`
	Status     ObjectStatus `json:"status"`
}
