# Go WASM Canvas Demo

This is a simple drawing application built using Go (compiled to WebAssembly) and the HTML Canvas API.

**Note:** This project serves as a technical demonstration and exploration of using Go with WebAssembly (`syscall/js` package) for HTML Canvas manipulation. It is not intended to be a feature-complete drawing application.

## Features

*   **Drawing Tools:**
    *   Pen (Freehand line drawing)
    *   Text (Editable directly on the canvas)
*   **Object Manipulation:**
    *   Select objects (lines, text)
    *   Move selected objects
    *   Scale selected objects proportionally (via control points)
    *   Delete selected objects (via button or Delete/Backspace key)
*   **Tool Switching:**
    *   Switch between Pen and Text tools using the toolbar buttons.

## Tech Stack

*   **Backend Logic:** Go (compiled to WebAssembly)
*   **Frontend:** HTML, CSS, JavaScript
*   **Drawing:** HTML Canvas API via `syscall/js`

## How to Run

1.  **Compile Go to WebAssembly:**
    ```bash
    GOOS=js GOARCH=wasm go build -o main.wasm
    ```
2.  **Serve the files:**
    You need a simple HTTP server to serve the `index.html`, `wasm_exec.js`, and `main.wasm` files. You can use any static file server. Here are a couple of options:

    *   **Using Go:**
        Create a `server.go` file:
        ```go
        // server.go
        package main

        import (
            "log"
            "net/http"
        )

        func main() {
            log.Println("Serving on http://localhost:8080")
            log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("."))))
        }
        ```
        Then run: `go run server.go`

    *   **Using Python:**
        ```bash
        # Python 3
        python -m http.server 8080
        # Python 2
        # python -m SimpleHTTPServer 8080
        ```
3.  **Open in Browser:**
    Navigate to `http://localhost:8080` in your web browser.

## Potential Future Exploration (Out of Scope for Demo)

*   Implement an Image object tool.
*   Improve the accuracy of text boundary calculations (e.g., using JS `measureText`).
*   Refine text editing interactions further.
*   Add more toolbar options (color picker, line width selection, etc.).
*   Implement robust error handling.
*   Optimize rendering performance for a large number of objects.