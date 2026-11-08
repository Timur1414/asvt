package repository

import (
	"math"
	"time"
)

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
}

func CalculateIoU(box1, box2 DetectedObject) float64 {
	interX1 := math.Max(float64(box1.X), float64(box2.X))
	interY1 := math.Max(float64(box1.Y), float64(box2.Y))
	interX2 := math.Min(float64(box1.X+box1.Width), float64(box2.X+box2.Width))
	interY2 := math.Min(float64(box1.Y+box1.Height), float64(box2.Y+box2.Height))

	interWidth := math.Max(0, interX2-interX1)
	interHeight := math.Max(0, interY2-interY1)
	interArea := interWidth * interHeight

	if interArea == 0 {
		return 0.0
	}
	area1 := float64(box1.Width * box1.Height)
	area2 := float64(box2.Width * box2.Height)

	unionArea := area1 + area2 - interArea
	if unionArea <= 0 {
		return 0.0
	}
	return interArea / unionArea
}
