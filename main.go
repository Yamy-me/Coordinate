package main

import (
	"fmt"
	"image/color"
	"math"
	"runtime"
	"slices"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/Knetic/govaluate"
)

type Points struct {
	X, Y float64
}

func main() {
	// Длина и ширина координатной плоскости
	// const height, width = 10, 10
	// CoordinateASCII := [height][width]string{}
	functions := []string{"x ** 2 + y ** 2 - 25", "(x + 2) ** 2 + (y - 2) ** 2 - 0.5", "(x - 2) ** 2 + (y - 2) ** 2 - 0.5", "0.2 * x ** 2 - 3"}
	
	points := []Points{}
	for _, v := range functions{
		result, _ := coordinateFnc(v)
		points = append(points, result...)
		
	}

	app := app.New()
	window := app.NewWindow("Название")
	window.Resize(fyne.NewSize(1920, 1080))

	// Горизонтальная линия
	centerX := float32(1920 / 2)
	vLine := canvas.NewLine(color.Black)
	vLine.Position1 = fyne.NewPos(centerX, 0)
	vLine.Position2 = fyne.NewPos(centerX, 1080)

	// Вертикальная линия
	centerY := float32(1080 / 2)
	hLine := canvas.NewLine(color.Black)
	hLine.Position1 = fyne.NewPos(0, centerY)
	hLine.Position2 = fyne.NewPos(1920, centerY)
	objectCanvas := []fyne.CanvasObject{hLine, vLine}

	// если функция не явная
	if slices.ContainsFunc(functions, func(v string) bool {
		return strings.Contains(v, "y")
	}){
		for _, p := range points {
		x := centerX + float32(p.X * 100)
		y := centerY - float32(p.Y * 100)
		dot := canvas.NewCircle(color.RGBA{255, 0, 0, 255})
		dot.Resize(fyne.NewSize(1, 1))
		dot.Move(fyne.NewPos(x, y))
		objectCanvas = append(objectCanvas, dot)
		}

	}else{
		for i := 1; i < len(points); i++ {
		x1 := centerX + float32(points[i-1].X * 100)
		y1 := centerY - float32(points[i-1].Y * 100)

		x2 := centerX + float32(points[i].X * 100)
		y2 := centerY - float32(points[i].Y * 100)
		
		Line := canvas.NewLine(color.RGBA{255, 0, 0, 255})
		Line.Position1 = fyne.NewPos(x1, y1)
		Line.Position2 = fyne.NewPos(x2, y2)
		objectCanvas = append(objectCanvas, Line)
		
		}
	}
	cont := container.NewWithoutLayout(objectCanvas...)
	// добавление в контейнер
	window.SetContent(cont)
	window.ShowAndRun()
	


}

func coordinateFnc(function string) ([]Points, error) {
	var wg sync.WaitGroup
	res := []Points{}
	fncs := map[string]govaluate.ExpressionFunction{
		"sin": func(args ...interface{}) (interface{}, error) {
			return math.Sin(args[0].(float64)), nil
		},
		"cos": func(args ...interface{}) (interface{}, error) {
			return math.Cos(args[0].(float64)), nil
		},
		"tan": func(args ...interface{}) (interface{}, error) {
			return math.Tan(args[0].(float64)), nil
		},
		"sqrt": func(args ...interface{}) (interface{}, error) {
			return math.Sqrt(args[0].(float64)), nil
		},
		"abs": func(args ...interface{}) (interface{}, error) {
			return math.Abs(args[0].(float64)), nil
		},
	}
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(function, fncs)
	if err != nil {
		return []Points{}, fmt.Errorf("Error in expr: NewEvaluableExpWithFnc: %w", err)
	}
	if strings.Contains(function, "y") {
		var channel = make(chan Points)
		workers := runtime.NumCPU()
		for i := range workers {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				step := (10.0 - -10.0) / float64(workers)
				xStart := -10.0 + float64(i)*step
				xEnd := -10.0 + float64((i+1))*step
				for x := xStart; x <= xEnd; x += 0.005 {
					for y := -10.0; y <= 10.0; y += 0.005 {
						params := map[string]interface{}{
							"x": x,
							"y": y,
						}
						result, err := expr.Evaluate(params)
						if err != nil {
							continue
						}
						if math.Abs(result.(float64)) < 0.005 {
							channel <- Points{X: x, Y: y}
						}
					}
				}
			}(i)
		}

		go func() {
			wg.Wait()
			close(channel)
		}()

		for v := range channel {
			res = append(res, Points{X: v.X, Y: v.Y})

		}

	} else {
		for x := -10.0; x <= 10.0; x += 0.01 {
			params := map[string]interface{}{
				"x": x,
			}
			y, err := expr.Evaluate(params)
			if err != nil {
				return []Points{}, fmt.Errorf("Error in Evaluate: %w", err)
			}
			res = append(res, Points{X: x, Y: y.(float64)})
		}
	}

	return res, nil
}
