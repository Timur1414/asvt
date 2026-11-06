package domain

type BoundingBox struct {
	XMin float64
	YMin float64
	XMax float64
	YMax float64
}

func (obj *BoundingBox) GetCenter() (float64, float64) {
	return (obj.XMin + obj.XMax) / 2, (obj.YMin + obj.YMax) / 2
}

func (obj *BoundingBox) GetSquare() float64 {
	return (obj.XMax - obj.XMin) * (obj.YMax - obj.YMin)
}

func (obj *BoundingBox) IsColliding(other *BoundingBox) bool {
	return false
}
