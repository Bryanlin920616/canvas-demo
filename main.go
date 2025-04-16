//go:build js && wasm

package main

import (
	"syscall/js"
)

func drawLine(this js.Value, args []js.Value) interface{} {
	// 接收兩點座標 (x1, y1, x2, y2)
	x1 := args[0].Float()
	y1 := args[1].Float()
	x2 := args[2].Float()
	y2 := args[3].Float()

	// 回傳一段簡單的線段資料 [x1, y1, x2, y2]
	result := js.Global().Get("Array").New(4)
	result.SetIndex(0, x1)
	result.SetIndex(1, y1)
	result.SetIndex(2, x2)
	result.SetIndex(3, y2)

	return result
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("drawLine", js.FuncOf(drawLine))

	<-c
}
