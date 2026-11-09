package domain

import "time"

type SecurityEvent struct {
	ObjectUuid       string
	ObjectType       string
	ObjectConfidence float32
	CreatedAt        time.Time
}
