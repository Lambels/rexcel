package main

import (
	"io"

	"github.com/xuri/excelize/v2"
)

type cell struct {
	id        uint
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
				id:        concat(uint(colx), uint(rowx)),
				val:       axis,
				isFormula: formula != "",
			}

		}
		colx = 0
	}

	return nil, nil
}

func concat(x, y uint) uint {
	var mul uint = 10
	for y >= mul {
		mul *= 10
	}
	return x*mul + y
}
