package main

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
)

type Points struct {
	X, Y float64
}

func main() {
	const height, width = 40, 40
	CoordinateASCII := [height][width]string{}
	points, err := coordinateFnc("(x ** 2 + y ** 2 - 1) ** 3 - x ** 2 * y ** 3")
	if err != nil{
		fmt.Println("error in main points section")
		return
	}
		for row := 0; row < height; row++ {
			for col := 0; col < width; col++ {
				CoordinateASCII[row][col] = " "
			}
		}
		midRow := height / 2
		midCol := width / 2

		// горизонтальная ось
		for col := 0; col < width; col++ {
			CoordinateASCII[midRow][col] = "-"
		}

		// вертикальная ось
		for row := 0; row < height; row++ {
			CoordinateASCII[row][midCol] = "|"
		}

		// пересечение
		CoordinateASCII[midRow][midCol] = "+"
		xMax, xMin := points[0].X, points[0].X
		yMax, yMin := points[0].Y, points[0].Y

		for _, v := range points {
			if v.X > xMax {
				xMax = v.X
			}
			if v.X < xMin {
				xMin = v.X
			}
			if v.Y > yMax {
				yMax = v.Y
			}
			if v.Y < yMin {
				yMin = v.Y
			}
		}

		for _, p := range points {
			col := int((p.X - xMin) / (xMax - xMin) * float64(width))
			row := int((yMax - p.Y) / (yMax - yMin) * float64(height))
			if row >= 0 && row < height && col >= 0 && col < width {
				CoordinateASCII[row][col] = "*"
			}
		}
		for _, row := range CoordinateASCII {
			fmt.Println(strings.Join(row[:], ""))
		}
	}

func coordinateFnc(function string) ([]Points, error) {
	var channel = make(chan Points)
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
		for i := range 20 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				step := (10.0 - -10.0) / 10
				xStart := -10.0 + float64(i)*step
				xEnd := -10.0 + float64((i+1))*step
				for x := xStart; x <= xEnd; x += 0.05 {
					for y := -10.0; y <= 10.0; y += 0.05{
						params := map[string]interface{}{
							"x": x,
							"y": y,
						}
						result, err := expr.Evaluate(params)
						if err != nil {
							continue
						}
						if math.Abs(result.(float64)) < 0.04 {
							channel <- Points{X: x, Y: y}
						}
					}
				}		
			}(i)
		}

		go func ()  {
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
