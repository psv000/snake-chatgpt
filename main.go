package main

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	window, err := initGlfw()
	panicOnError(err)
	defer glfw.Terminate()

	program, err := newProgram(vertexShaderSource, fragmentShaderSource)
	panicOnError(err)

	colorLoc := location(program, "color")
	modelLoc := location(program, "model")
	projectionLoc := location(program, "projection")

	vao := createVAOAndVBO(squareVertices, indices)

	s := NewSnake(windowWidth, windowHeight)
	f := NewFood(s, windowWidth, windowHeight)

	for !window.ShouldClose() {
		processInput(window, s)
		//s.move()
		gl.Clear(gl.COLOR_BUFFER_BIT)

		s.Draw(program, vao)
		f.Draw(program, vao, float32(windowWidth), float32(windowHeight))

		glfw.PollEvents()
		window.SwapBuffers()
	}
}
