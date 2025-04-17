//go:build js && wasm

package canvas

import (
	"math"
	"syscall/js"

	"canvas-demo/internal/canvas/shape"
)

// CanvasManager 處理所有 Canvas 相關操作
type CanvasManager struct {
	canvas        js.Value
	ctx           js.Value
	width         float64
	height        float64
	shapes        []shape.Shape
	currentLine   *shape.Line
	selectedShape shape.Shape
	isDragging    bool
	isScaling     bool
	activeControl shape.ControlPoint
	lastX         float64
	lastY         float64
	scaleCenter   shape.Point
}

// NewCanvasManager 創建新的 Canvas 管理器
func NewCanvasManager(canvasID string) *CanvasManager {
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", canvasID)
	ctx := canvas.Call("getContext", "2d")

	return &CanvasManager{
		canvas: canvas,
		ctx:    ctx,
		width:  canvas.Get("width").Float(),
		height: canvas.Get("height").Float(),
		shapes: make([]shape.Shape, 0),
	}
}

// Clear 清除整個畫布
func (cm *CanvasManager) Clear() {
	cm.ctx.Call("clearRect", 0, 0, cm.width, cm.height)
}

// SetStrokeStyle 設置線條樣式
func (cm *CanvasManager) SetStrokeStyle(color string) {
	cm.ctx.Set("strokeStyle", color)
}

// SetLineWidth 設置線條寬度
func (cm *CanvasManager) SetLineWidth(width float64) {
	cm.ctx.Set("lineWidth", width)
}

// StartDrawing 開始繪圖
func (cm *CanvasManager) StartDrawing(x, y float64) {
	p := shape.Point{X: x, Y: y}

	// 如果有選中的形狀，檢查是否點擊到控制點
	if cm.selectedShape != nil {
		controlPoint := cm.selectedShape.HitControl(p)
		if controlPoint != shape.None {
			cm.isScaling = true
			cm.activeControl = controlPoint
			cm.scaleCenter = cm.getScaleCenter(controlPoint)
			cm.lastX = x
			cm.lastY = y
			return
		}
	}

	// 檢查是否點擊到現有形狀
	clickedShape := cm.findShapeAt(p)
	if clickedShape != nil {
		if cm.selectedShape != nil {
			cm.setSelectedShape(nil)
		}
		cm.setSelectedShape(clickedShape)
		cm.isDragging = true
		cm.lastX = x
		cm.lastY = y
		return
	}

	// 如果沒有點擊到形狀，取消當前選中
	if cm.selectedShape != nil {
		cm.setSelectedShape(nil)
	}

	// 開始新的線段
	cm.currentLine = shape.NewLine(shape.Style{
		StrokeStyle: cm.ctx.Get("strokeStyle").String(),
		LineWidth:   cm.ctx.Get("lineWidth").Float(),
	})
	cm.currentLine.AddPoint(p)
}

// Draw 繪製
func (cm *CanvasManager) Draw(x, y float64) {
	if cm.isScaling && cm.selectedShape != nil {
		// 處理縮放
		// 計算主要縮放方向
		var scale float64
		switch cm.activeControl {
		case shape.TopLeft:
			// 使用對角線距離變化來計算縮放比例
			currentDist := math.Sqrt(math.Pow(x-cm.scaleCenter.X, 2) + math.Pow(y-cm.scaleCenter.Y, 2))
			originalDist := math.Sqrt(math.Pow(cm.lastX-cm.scaleCenter.X, 2) + math.Pow(cm.lastY-cm.scaleCenter.Y, 2))
			scale = currentDist / originalDist
			// 當距離增加時應該縮小，反之應該放大
			scale = 1 / scale
		case shape.TopRight:
			currentDist := math.Sqrt(math.Pow(x-cm.scaleCenter.X, 2) + math.Pow(y-cm.scaleCenter.Y, 2))
			originalDist := math.Sqrt(math.Pow(cm.lastX-cm.scaleCenter.X, 2) + math.Pow(cm.lastY-cm.scaleCenter.Y, 2))
			scale = currentDist / originalDist
			// 當距離增加時應該放大
			if x < cm.scaleCenter.X {
				scale = 1 / scale
			}
		case shape.BottomLeft:
			currentDist := math.Sqrt(math.Pow(x-cm.scaleCenter.X, 2) + math.Pow(y-cm.scaleCenter.Y, 2))
			originalDist := math.Sqrt(math.Pow(cm.lastX-cm.scaleCenter.X, 2) + math.Pow(cm.lastY-cm.scaleCenter.Y, 2))
			scale = currentDist / originalDist
			// 當距離增加時應該放大
			if x > cm.scaleCenter.X {
				scale = 1 / scale
			}
		case shape.BottomRight:
			currentDist := math.Sqrt(math.Pow(x-cm.scaleCenter.X, 2) + math.Pow(y-cm.scaleCenter.Y, 2))
			originalDist := math.Sqrt(math.Pow(cm.lastX-cm.scaleCenter.X, 2) + math.Pow(cm.lastY-cm.scaleCenter.Y, 2))
			scale = currentDist / originalDist
		}

		// 限制最小縮放比例
		const minScale = 0.1
		if scale < minScale {
			scale = minScale
		}

		// 使用相同的縮放比例進行等比例縮放
		cm.selectedShape.Scale(scale, scale, cm.scaleCenter)
		cm.lastX = x
		cm.lastY = y
		cm.redraw()
		return
	}

	if cm.isDragging && cm.selectedShape != nil {
		// 移動選中的形狀
		dx := x - cm.lastX
		dy := y - cm.lastY
		cm.selectedShape.Move(dx, dy)
		cm.lastX = x
		cm.lastY = y
		cm.redraw()
		return
	}

	if cm.currentLine != nil {
		// 繼續繪製當前線段
		cm.currentLine.AddPoint(shape.Point{X: x, Y: y})
		cm.redraw()
	}
}

// StopDrawing 停止繪圖
func (cm *CanvasManager) StopDrawing() {
	if cm.isScaling {
		cm.isScaling = false
		cm.activeControl = shape.None
		return
	}

	if cm.isDragging {
		cm.isDragging = false
		return
	}

	if cm.currentLine != nil {
		cm.shapes = append(cm.shapes, cm.currentLine)
		cm.currentLine = nil
	}
}

// DeleteSelected 刪除選中的形狀
func (cm *CanvasManager) DeleteSelected() {
	if cm.selectedShape == nil {
		return
	}

	// 從形狀列表中移除
	for i, s := range cm.shapes {
		if s == cm.selectedShape {
			cm.shapes = append(cm.shapes[:i], cm.shapes[i+1:]...)
			break
		}
	}

	cm.selectedShape = nil
	cm.redraw()
}

// setSelectedShape 設置選中的形狀
func (cm *CanvasManager) setSelectedShape(s shape.Shape) {
	// 取消之前選中形狀的選中狀態
	if cm.selectedShape != nil {
		if line, ok := cm.selectedShape.(*shape.Line); ok {
			line.SetSelected(false)
		}
	}

	cm.selectedShape = s

	// 設置新形狀的選中狀態
	if s != nil {
		if line, ok := s.(*shape.Line); ok {
			line.SetSelected(true)
		}
	}

	cm.redraw()
}

// getScaleCenter 根據控制點獲取縮放中心
func (cm *CanvasManager) getScaleCenter(cp shape.ControlPoint) shape.Point {
	bounds := cm.selectedShape.GetBounds()
	switch cp {
	case shape.TopLeft:
		return shape.Point{
			X: bounds.X + bounds.Width,
			Y: bounds.Y + bounds.Height,
		}
	case shape.TopRight:
		return shape.Point{
			X: bounds.X,
			Y: bounds.Y + bounds.Height,
		}
	case shape.BottomLeft:
		return shape.Point{
			X: bounds.X + bounds.Width,
			Y: bounds.Y,
		}
	case shape.BottomRight:
		return shape.Point{
			X: bounds.X,
			Y: bounds.Y,
		}
	default:
		return shape.Point{
			X: bounds.X + bounds.Width/2,
			Y: bounds.Y + bounds.Height/2,
		}
	}
}

// GetMousePosition 獲取滑鼠在 Canvas 上的位置
func (cm *CanvasManager) GetMousePosition(event js.Value) (float64, float64) {
	rect := cm.canvas.Call("getBoundingClientRect")
	x := event.Get("clientX").Float() - rect.Get("left").Float()
	y := event.Get("clientY").Float() - rect.Get("top").Float()
	return x, y
}

// findShapeAt 找到指定位置的形狀
func (cm *CanvasManager) findShapeAt(p shape.Point) shape.Shape {
	// 從後往前檢查，這樣可以選中最上層的形狀
	for i := len(cm.shapes) - 1; i >= 0; i-- {
		if cm.shapes[i].Contains(p) {
			return cm.shapes[i]
		}
	}
	return nil
}

// redraw 重新繪製所有形狀
func (cm *CanvasManager) redraw() {
	cm.Clear()

	// 繪製所有已完成的形狀
	for _, s := range cm.shapes {
		s.Draw(cm.ctx)
	}

	// 繪製當前正在繪製的線段
	if cm.currentLine != nil {
		cm.currentLine.Draw(cm.ctx)
	}
}
