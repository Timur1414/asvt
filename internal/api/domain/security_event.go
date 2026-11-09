package domain

import "time"

type SecurityEvent struct {
	Id               int
	ObjectUuid       string
	ObjectType       string
	ObjectConfidence float32
	CreatedAt        time.Time
}
