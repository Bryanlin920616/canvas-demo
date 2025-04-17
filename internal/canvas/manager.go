//go:build js && wasm

package canvas

import (
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
	lastX         float64
	lastY         float64
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
	// 檢查是否點擊到現有形狀
	clickedShape := cm.findShapeAt(shape.Point{X: x, Y: y})
	if clickedShape != nil {
		cm.selectedShape = clickedShape
		cm.isDragging = true
		cm.lastX = x
		cm.lastY = y
		return
	}

	// 如果沒有點擊到形狀，開始新的線段
	cm.currentLine = shape.NewLine(shape.Style{
		StrokeStyle: cm.ctx.Get("strokeStyle").String(),
		LineWidth:   cm.ctx.Get("lineWidth").Float(),
	})
	cm.currentLine.AddPoint(shape.Point{X: x, Y: y})
}

// Draw 繪製
func (cm *CanvasManager) Draw(x, y float64) {
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
	if cm.isDragging {
		cm.isDragging = false
		cm.selectedShape = nil
		return
	}

	if cm.currentLine != nil {
		cm.shapes = append(cm.shapes, cm.currentLine)
		cm.currentLine = nil
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
