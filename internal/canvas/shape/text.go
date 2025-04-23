//go:build js && wasm

package shape

import (
	"fmt"
	"syscall/js"
)

// Text 表示文字物件
type Text struct {
	Content    string
	Position   Point
	Style      TextStyle
	isSelected bool
	isEditing  bool     // 是否正在編輯
	inputElem  js.Value // HTML input 元素
}

// TextStyle 定義文字的樣式
type TextStyle struct {
	Font      string  // 例如："20px Arial"
	FillStyle string  // 文字顏色
	Size      float64 // 字體大小（像素）
}

// NewText 創建新的文字物件
func NewText(position Point, textStyle TextStyle) *Text {
	// 創建一個輸入框元素
	doc := js.Global().Get("document")
	input := doc.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("value", "新文字")

	// 設置樣式
	inputStyle := input.Get("style")
	inputStyle.Set("position", "fixed") // 改用 fixed 定位
	inputStyle.Set("font", textStyle.Font)
	inputStyle.Set("color", textStyle.FillStyle)
	inputStyle.Set("border", "1px solid #ccc") // 添加邊框以便於識別
	inputStyle.Set("padding", "2px 4px")       // 添加內邊距
	inputStyle.Set("outline", "none")
	inputStyle.Set("background", "white") // 設置背景色
	inputStyle.Set("display", "none")
	inputStyle.Set("z-index", "1000")   // 確保在畫布上層
	inputStyle.Set("min-width", "50px") // 最小寬度
	inputStyle.Set("cursor", "text")    // 文字游標

	// 將輸入框添加到文檔中
	doc.Get("body").Call("appendChild", input)

	text := &Text{
		Content:   "新文字",
		Position:  position,
		Style:     textStyle,
		inputElem: input,
	}

	// 添加事件監聽器
	input.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		args[0].Call("stopPropagation") // 阻止冒泡到 Canvas
		return nil
	}))

	input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		text.Content = input.Get("value").String()
		return nil
	}))

	// 阻止 Delete 和 Backspace 鍵冒泡到 document
	input.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		key := event.Get("key").String()
		if key == "Delete" || key == "Backspace" {
			event.Call("stopPropagation")
		}
		return nil
	}))

	return text
}

// StartEditing 開始編輯文字
func (t *Text) StartEditing(canvasRect js.Value) {
	if !t.isEditing {
		t.isEditing = true

		// 計算輸入框位置
		left := t.Position.X + canvasRect.Get("left").Float()
		top := t.Position.Y + canvasRect.Get("top").Float() - t.Style.Size

		// 設置輸入框位置和樣式
		style := t.inputElem.Get("style")
		style.Set("left", fmt.Sprintf("%dpx", int(left)))
		style.Set("top", fmt.Sprintf("%dpx", int(top)))
		style.Set("font", t.Style.Font)
		style.Set("display", "block")

		t.inputElem.Set("value", t.Content)

		// 聚焦並選中全部文字
		t.inputElem.Call("focus")
		t.inputElem.Call("select")
	}
}

// StopEditing 停止編輯文字
func (t *Text) StopEditing() {
	if t.isEditing {
		t.isEditing = false
		t.inputElem.Get("style").Set("display", "none")
	}
}

// Delete 刪除文字
func (t *Text) Delete() {
	// 移除輸入框元素
	t.inputElem.Call("remove")
}

// Draw 繪製文字
func (t *Text) Draw(ctx js.Value) {
	if !t.isEditing {
		ctx.Set("font", t.Style.Font)
		ctx.Set("fillStyle", t.Style.FillStyle)

		// 繪製文字
		ctx.Call("fillText", t.Content, t.Position.X, t.Position.Y)
	}

	// 如果被選中，繪製控制點
	if t.isSelected {
		t.DrawControls(ctx)
	}
}

// Contains 檢查點是否在文字範圍內
func (t *Text) Contains(p Point) bool {
	if t.isEditing {
		return false // 編輯時不處理選中
	}
	bounds := t.GetBounds()
	return p.X >= bounds.X && p.X <= bounds.X+bounds.Width &&
		p.Y >= bounds.Y && p.Y <= bounds.Y+bounds.Height
}

// Move 移動文字
func (t *Text) Move(dx, dy float64) {
	t.Position.X += dx
	t.Position.Y += dy
}

// GetBounds 獲取文字的邊界
func (t *Text) GetBounds() Bounds {
	// 這裡需要使用 JavaScript 的 measureText 來獲取文字的實際尺寸
	// 由於無法直接在這裡訪問 context，我們使用一個估算值
	width := float64(len(t.Content)) * t.Style.Size * 0.6 // 估算寬度
	height := t.Style.Size * 1.2                          // 估算高度

	return Bounds{
		X:      t.Position.X,
		Y:      t.Position.Y - height + 5, // 上移一點，因為 Y 座標是文字的基線位置
		Width:  width,
		Height: height,
	}
}

// Scale 縮放文字
func (t *Text) Scale(sx, sy float64, center Point) {
	// 相對於中心點進行縮放
	dx := t.Position.X - center.X
	dy := t.Position.Y - center.Y

	t.Position.X = center.X + dx*sx
	t.Position.Y = center.Y + dy*sy

	// 更新字體大小
	t.Style.Size *= sx
	t.Style.Font = fmt.Sprintf("%.0fpx Arial", t.Style.Size)
}

// DrawControls 繪製控制點
func (t *Text) DrawControls(ctx js.Value) {
	bounds := t.GetBounds()

	// 保存當前繪圖狀態
	ctx.Call("save")

	// 設置控制點樣式
	ctx.Set("fillStyle", "#ffffff")
	ctx.Set("strokeStyle", "#000000")
	ctx.Set("lineWidth", 1)

	// 繪製邊界框
	ctx.Call("beginPath")
	ctx.Call("rect", bounds.X, bounds.Y, bounds.Width, bounds.Height)
	ctx.Call("stroke")

	const controlSize = 5.0 // 控制點大小

	// 繪製控制點
	controlPoints := []struct {
		x, y float64
	}{
		{bounds.X, bounds.Y},                                // TopLeft
		{bounds.X + bounds.Width, bounds.Y},                 // TopRight
		{bounds.X, bounds.Y + bounds.Height},                // BottomLeft
		{bounds.X + bounds.Width, bounds.Y + bounds.Height}, // BottomRight
	}

	for _, cp := range controlPoints {
		ctx.Call("beginPath")
		ctx.Call("arc", cp.x, cp.y, controlSize, 0, 2*3.14159, false)
		ctx.Call("fill")
		ctx.Call("stroke")
	}

	// 恢復繪圖狀態
	ctx.Call("restore")
}

// HitControl 檢查是否點擊到控制點
func (t *Text) HitControl(p Point) ControlPoint {
	bounds := t.GetBounds()
	const controlSize = 5.0         // 視覺上的控制點大小
	const hitArea = controlSize * 4 // 增加點選範圍到視覺大小的4倍

	// 檢查各個控制點
	controlPoints := []struct {
		point ControlPoint
		x, y  float64
	}{
		{TopLeft, bounds.X, bounds.Y},
		{TopRight, bounds.X + bounds.Width, bounds.Y},
		{BottomLeft, bounds.X, bounds.Y + bounds.Height},
		{BottomRight, bounds.X + bounds.Width, bounds.Y + bounds.Height},
	}

	for _, cp := range controlPoints {
		if distance(p, Point{X: cp.x, Y: cp.y}) <= hitArea {
			return cp.point
		}
	}

	return None
}

// SetSelected 設置選中狀態
func (t *Text) SetSelected(selected bool) {
	t.isSelected = selected
}
