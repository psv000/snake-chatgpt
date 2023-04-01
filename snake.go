package main

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Point struct {
	X, Y float32
}

const gridSize = 20

func drawSquare(program, vao uint32, indices []uint32, x, y, winWidth, winHeight float32) {
	// Выберите шейдерную программу и VAO
	gl.UseProgram(program)
	gl.BindVertexArray(vao)

	// Вычислить модель-вид-проекцию матрицы для смещения квадрата
	model := mgl32.Translate3D(x, y, 0)
	view := mgl32.Ident4()
	projection := mgl32.Ortho2D(0, winWidth, winHeight, 0)
	mvp := projection.Mul4(view).Mul4(model)

	// Установить MVP матрицу в шейдере
	mvpUniform := gl.GetUniformLocation(program, gl.Str("mvp\x00"))
	gl.UniformMatrix4fv(mvpUniform, 1, false, &mvp[0])

	// Рисовать квадрат
	gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, nil)

	// Отвязать VAO и программу
	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

type Snake struct {
	Body    []Point
	Dir     Direction
	Dead    bool
	Grow    bool
	WinSize [2]float32
}

func NewSnake(winWidth, winHeight float32) *Snake {
	snake := &Snake{
		Body: []Point{
			{winWidth / 2, winHeight / 2},
			{winWidth/2 - gridSize, winHeight / 2},
		},
		Dir:     Right,
		Dead:    false,
		Grow:    false,
		WinSize: [2]float32{winWidth, winHeight},
	}
	return snake
}

func (s *Snake) Draw(program, vao uint32) {
	for _, p := range s.Body {
		drawSquare(program, vao, indices, p.X, p.Y, s.WinSize[0], s.WinSize[1])
	}
}

func (s *Snake) move() {
	if s.Grow {
		s.Grow = false
	} else {
		s.Body = s.Body[:len(s.Body)-1]
	}

	newHead := s.Body[0]
	switch s.Dir {
	case Up:
		newHead.Y += gridSize
	case Down:
		newHead.Y -= gridSize
	case Left:
		newHead.X -= gridSize
	case Right:
		newHead.X += gridSize
	}

	if newHead.X < 0 || newHead.X >= s.WinSize[0] || newHead.Y < 0 || newHead.Y >= s.WinSize[1] {
		s.Dead = true
		return
	}

	for _, p := range s.Body {
		if newHead.X == p.X && newHead.Y == p.Y {
			s.Dead = true
			return
		}
	}

	s.Body = append([]Point{newHead}, s.Body...)
}

func processInput(window *glfw.Window, snake *Snake) {
	if window.GetKey(glfw.KeyUp) == glfw.Press && snake.Dir != Down {
		snake.Dir = Up
	} else if window.GetKey(glfw.KeyDown) == glfw.Press && snake.Dir != Up {
		snake.Dir = Down
	} else if window.GetKey(glfw.KeyLeft) == glfw.Press && snake.Dir != Right {
		snake.Dir = Left
	} else if window.GetKey(glfw.KeyRight) == glfw.Press && snake.Dir != Left {
		snake.Dir = Right
	}
}
