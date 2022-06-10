package main

import (
	"io"
	"math"

	"github.com/xuri/excelize/v2"
)

type cell struct {
	id        int
	lowLink   int
	val       string
	isFormula bool
}

type digraph struct {
	formulas []*cell
	cells    []*cell
	edges    map[cell][]*cell
}

func newGraph(from io.Reader) (*digraph, error) {
	f, err := excelize.OpenReader(from)
	if err != nil {
		return nil, err
	}

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	var colx, rowx int
	for _, row := range rows {
		rowx++
		for _, colCell := range row {
			colx++
			if colCell == "" {
				continue
			}

			axis, err := excelize.CoordinatesToCellName(colx, rowx)
			if err != nil {
				return nil, err
			}

			formula, err := f.GetCellFormula("Sheet1", axis)
			if err != nil {
				return nil, err
			}

			c := &cell{
				id:        idFromPoint(colx, rowx),
				val:       axis,
				isFormula: formula != "",
			}

		}
		colx = 0
	}

	return nil, nil
}

func idFromPoint(x, y int) int {
	var nTmp int

	// calc length of x, y ie: (123 -> 3).
	var lenX int
	nTmp = x
	for nTmp > 0 {
		nTmp = nTmp / 10
		lenX++
	}

	var lenY int
	nTmp = y
	for nTmp > 0 {
		nTmp = nTmp / 10
		lenY++
	}

	digitsX, digitsY := calcDigit(x, lenX, 1, make([]int, lenX)), calcDigit(y, lenY, 1, make([]int, lenY))

	// add digits to sum with padding.
	var finalNum int
	for i, v := range digitsX {
		finalNum += v * (int(math.Pow10(i + len(digitsY))))
	}

	for i, v := range digitsY {
		finalNum += v * (int(math.Pow10(i)))
	}

	return finalNum
}

func calcDigit(n, length, depth int, final []int) []int {
	if depth == length+1 {
		return final
	}

	// get most significant digit.
	var digit int
	nTmp := n
	for nTmp >= 10 {
		nTmp = nTmp / 10
	}
	digit = nTmp

	// subtract it with its relative depth to the length.
	n = n - (digit * int(math.Pow10(length-depth)))
	final[depth-1] = digit

	return calcDigit(n, length, depth+1, final)
}
