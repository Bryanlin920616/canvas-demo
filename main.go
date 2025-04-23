//go:build js && wasm

package main

import (
	"syscall/js"

	"canvas-demo/internal/canvas"
)

var canvasManager *canvas.CanvasManager

func main() {
	c := make(chan struct{})

	// 初始化 Canvas 管理器
	canvasManager = canvas.NewCanvasManager("canvas")

	// 設置預設樣式
	canvasManager.SetStrokeStyle("#000000")
	canvasManager.SetLineWidth(2)

	// 註冊事件處理函數
	js.Global().Set("startDrawing", js.FuncOf(startDrawing))
	js.Global().Set("drawing", js.FuncOf(drawing))
	js.Global().Set("stopDrawing", js.FuncOf(stopDrawing))
	js.Global().Set("deleteSelectedShape", js.FuncOf(deleteSelectedShape))
	js.Global().Set("setCurrentTool", js.FuncOf(setCurrentTool))

	<-c
}

func startDrawing(this js.Value, args []js.Value) interface{} {
	event := args[0]
	x, y := canvasManager.GetMousePosition(event)
	canvasManager.StartDrawing(x, y)
	return nil
}

func drawing(this js.Value, args []js.Value) interface{} {
	event := args[0]
	x, y := canvasManager.GetMousePosition(event)
	canvasManager.Draw(x, y)
	return nil
}

func stopDrawing(this js.Value, args []js.Value) interface{} {
	canvasManager.StopDrawing()
	return nil
}

func deleteSelectedShape(this js.Value, args []js.Value) interface{} {
	canvasManager.DeleteSelected()
	return nil
}

func setCurrentTool(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		tool := args[0].String()
		canvasManager.SetCurrentTool(tool)
	}
	return nil
}
