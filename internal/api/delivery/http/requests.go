package http

type BBoxRequest struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type DetectedObjectRequest struct {
	ClassId    int         `json:"class_id"`
	ClassName  string      `json:"class_name"`
	Confidence float32     `json:"confidence"`
	Bbox       BBoxRequest `json:"bbox"`
}

type DetectRequest struct {
	DeviceID    string                  `json:"device_id"`
	Timestamp   int                     `json:"timestamp"`
	TimestampMs int                     `json:"timestamp_ms"`
	FrameId     int                     `json:"frame_id"`
	FrameWidth  int                     `json:"frame_width"`
	FrameHeight int                     `json:"frame_height"`
	Objects     []DetectedObjectRequest `json:"objects"`
}
