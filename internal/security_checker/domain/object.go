package domain

import "time"

type DetectedObject struct {
	Uuid       string    `json:"uuid"`
	ClassId    int       `json:"class_id"`
	ClassName  string    `json:"class_name"`
	Confidence float32   `json:"confidence"`
	X          int       `json:"x"`
	Y          int       `json:"y"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	CreatedAt  time.Time `json:"created_at"`
	Noted      bool      `json:"noted"`
}
