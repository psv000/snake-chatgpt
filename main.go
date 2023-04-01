package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width              = 640
	height             = 480
	rows               = 40
	cols               = int(40. / (float64(width) / float64(height)))
	snakeSquareSize    = width / rows
	foodSquareSize     = width / rows
	snakeInitialLength = 5
)

var (
	vertexShaderSource = `
		#version 410 core
		layout (location = 0) in vec2 aPos;
		uniform mat4 model;
		uniform mat4 projection;

		void main() {
			gl_Position = projection * model * vec4(aPos, 0.0, 1.0);
		}
	` + "\x00"

	fragmentShaderSource = `
		#version 410 core
		out vec4 FragColor;
		uniform vec3 color;

		void main() {
			FragColor = vec4(color, 1.0);
		}
	` + "\x00"

	squareVertices = []float32{
		0, 0,
		1, 0,
		1, 1,
		0, 1,
	}

	indices = []uint32{
		0, 1, 2,
		0, 2, 3,
	}
)

type point struct {
	x, y int
}

type snakePart struct {
	position point
}

type food struct {
	position point
}

var (
	snake         []snakePart
	snakeVelocity point
	foodInstance  food
)

func init() {
	rand.Seed(time.Now().UnixNano())
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		panic(fmt.Errorf("could not initialize glfw: %v", err))
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Snake", nil, nil)
	if err != nil {
		panic(fmt.Errorf("could not create opengl renderer: %v", err))
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(fmt.Errorf("could not initialize OpenGL: %v", err))
	}

	gl.Viewport(0, 0, width, height)

	fmt.Printf("gl version %s\n", gl.GoStr(gl.GetString(gl.VERSION)))
	fmt.Printf("render version %s\n", gl.GoStr(gl.GetString(gl.RENDERER)))

	// Shaders
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	program, err := createProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// VBO, VAO, EBO
	var VBO, VAO, EBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &EBO)

	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(squareVertices)*4, gl.Ptr(squareVertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	initSnake()
	createFood()

	colorLoc := gl.GetUniformLocation(program, gl.Str("color\x00"))
	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))

	for !window.ShouldClose() {
		processInput(window)
		update()

		checkGLError()

		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(program)

		projection := mgl32.Ortho2D(0, float32(width), float32(height), 0)

		gl.UniformMatrix4fv(projectionLoc, 1, false, &projection[0])

		renderSnake(colorLoc, modelLoc, VAO)
		renderFood(colorLoc, modelLoc, VAO)

		glfw.PollEvents()
		window.SwapBuffers()

		time.Sleep(75 * time.Millisecond)
	}
}

func initSnake() {
	snake = make([]snakePart, snakeInitialLength)
	for i := 0; i < snakeInitialLength; i++ {
		snake[i] = snakePart{point{rows/2 - i, cols / 2}}
	}
	snakeVelocity = point{1, 0}
}

func createFood() {
	x, y := rand.Intn(rows), rand.Intn(cols)
	fmt.Printf("create food x: %d y %d\n", x, y)
	foodInstance = food{point{x, y}}
}

func update() {
	// Update snake position
	head := snake[0].position
	newHead := point{head.x + snakeVelocity.x, head.y + snakeVelocity.y}

	// Check for collision with wall
	if newHead.x < 0 || newHead.x >= rows || newHead.y < 0 || newHead.y >= cols {
		initSnake()
		return
	}

	// Check for collision with self
	for _, part := range snake {
		if newHead.x == part.position.x && newHead.y == part.position.y {
			initSnake()
			return
		}
	}

	// Check for collision with food
	if newHead.x == foodInstance.position.x && newHead.y == foodInstance.position.y {
		// Add new snake part
		newPart := snakePart{point{newHead.x, newHead.y}}
		snake = append([]snakePart{newPart}, snake...)

		// Create new food
		createFood()
	}

	// Move snake
	for i := len(snake) - 1; i > 0; i-- {
		snake[i].position = snake[i-1].position
	}
	snake[0].position = newHead
}

func renderSnake(colorLoc int32, modelLoc int32, VAO uint32) {
	gl.Uniform3f(colorLoc, 0.0, 1.0, 0.0)
	gl.BindVertexArray(VAO)

	for _, part := range snake {
		x, y := float32(part.position.x*snakeSquareSize), float32(part.position.y*snakeSquareSize)
		model := mgl32.Translate3D(x, y, 0).Mul4(mgl32.Scale3D(float32(snakeSquareSize), float32(snakeSquareSize), 1))

		checkGLError()

		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

		gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, nil)
	}

	gl.BindVertexArray(0)
}

func renderFood(colorLoc int32, modelLoc int32, VAO uint32) {
	gl.Uniform3f(colorLoc, 1.0, 0.0, 0.0)
	gl.BindVertexArray(VAO)

	x, y := float32(foodInstance.position.x*foodSquareSize), float32(foodInstance.position.y*foodSquareSize)
	model := mgl32.Translate3D(x, y, 0).Mul4(mgl32.Scale3D(float32(foodSquareSize), float32(foodSquareSize), 1.))

	gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

	gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, nil)

	gl.BindVertexArray(0)
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])

		return 0, fmt.Errorf("failed to compile %v: %v", source, string(log))
	}

	return shader, nil
}

func createProgram(vertexShader uint32, fragmentShader uint32) (uint32, error) {
	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(program, logLength, nil, &log[0])

		return 0, fmt.Errorf("failed to link program: %v", string(log))
	}

	gl.DetachShader(program, vertexShader)
	gl.DetachShader(program, fragmentShader)

	return program, nil
}

func processInput(window *glfw.Window) {
	if window.GetKey(glfw.KeyW) == glfw.Press && snakeVelocity.y != 1 {
		snakeVelocity.x = 0
		snakeVelocity.y = -1
	} else if window.GetKey(glfw.KeyS) == glfw.Press && snakeVelocity.y != -1 {
		snakeVelocity.x = 0
		snakeVelocity.y = 1
	} else if window.GetKey(glfw.KeyA) == glfw.Press && snakeVelocity.x != 1 {
		snakeVelocity.x = -1
		snakeVelocity.y = 0
	} else if window.GetKey(glfw.KeyD) == glfw.Press && snakeVelocity.x != -1 {
		snakeVelocity.x = 1
		snakeVelocity.y = 0
	}
}

func checkGLError() {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		fmt.Printf("OpenGL error: %v\n", err)
	}
}
