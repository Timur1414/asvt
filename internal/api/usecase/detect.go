package usecase

type DetectUseCase interface {
	// ToDo get nearest object (by Bbox)
	// ToDo save new objects + create uuid
}

type Detect struct {
}

func NewDetect() *Detect {
	return &Detect{}
}
