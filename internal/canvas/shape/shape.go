//go:build js && wasm

package shape

import (
	"syscall/js"
)

// Point 表示座標點
type Point struct {
	X float64
	Y float64
}

// Shape 定義基本形狀介面
type Shape interface {
	Draw(ctx js.Value)
	Contains(p Point) bool
	Move(dx, dy float64)
	GetBounds() Bounds
}

// Bounds 表示形狀的邊界
type Bounds struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Line 表示線段
type Line struct {
	Points []Point
	Style  Style
}

// Style 定義形狀的樣式
type Style struct {
	StrokeStyle string
	LineWidth   float64
}

// NewLine 創建新的線段
func NewLine(style Style) *Line {
	return &Line{
		Points: make([]Point, 0),
		Style:  style,
	}
}

// AddPoint 添加點到線段
func (l *Line) AddPoint(p Point) {
	l.Points = append(l.Points, p)
}

// Draw 繪製線段
func (l *Line) Draw(ctx js.Value) {
	if len(l.Points) < 2 {
		return
	}

	ctx.Set("strokeStyle", l.Style.StrokeStyle)
	ctx.Set("lineWidth", l.Style.LineWidth)

	ctx.Call("beginPath")
	ctx.Call("moveTo", l.Points[0].X, l.Points[0].Y)

	for i := 1; i < len(l.Points); i++ {
		ctx.Call("lineTo", l.Points[i].X, l.Points[i].Y)
	}

	ctx.Call("stroke")
}

// Contains 檢查點是否在線段上
func (l *Line) Contains(p Point) bool {
	const threshold = 5.0 // 選取容差

	for i := 1; i < len(l.Points); i++ {
		p1 := l.Points[i-1]
		p2 := l.Points[i]

		// 計算點到線段的距離
		d := pointToLineDistance(p, p1, p2)
		if d <= threshold {
			return true
		}
	}
	return false
}

// Move 移動線段
func (l *Line) Move(dx, dy float64) {
	for i := range l.Points {
		l.Points[i].X += dx
		l.Points[i].Y += dy
	}
}

// GetBounds 獲取線段的邊界
func (l *Line) GetBounds() Bounds {
	if len(l.Points) == 0 {
		return Bounds{}
	}

	minX := l.Points[0].X
	minY := l.Points[0].Y
	maxX := l.Points[0].X
	maxY := l.Points[0].Y

	for _, p := range l.Points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	return Bounds{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// pointToLineDistance 計算點到線段的距離
func pointToLineDistance(p, start, end Point) float64 {
	// 使用向量計算點到線段的距離
	l2 := (end.X-start.X)*(end.X-start.X) + (end.Y-start.Y)*(end.Y-start.Y)
	if l2 == 0 {
		// start 和 end 是同一點
		return distance(p, start)
	}

	t := ((p.X-start.X)*(end.X-start.X) + (p.Y-start.Y)*(end.Y-start.Y)) / l2

	if t < 0 {
		return distance(p, start)
	}
	if t > 1 {
		return distance(p, end)
	}

	return distance(p, Point{
		X: start.X + t*(end.X-start.X),
		Y: start.Y + t*(end.Y-start.Y),
	})
}

// distance 計算兩點之間的距離
func distance(p1, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return (dx*dx + dy*dy)
}
