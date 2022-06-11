package main

import (
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

var validRunes = [...]bool{
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'(': true,
	')': true,
}

type cell struct {
	id      uint
	y       uint
	lowLink int
	val     string
}

type digraph struct {
	formulas map[uint]*cell
	cells    map[uint]*cell
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

	graph := &digraph{
		formulas: make(map[uint]*cell),
		cells:    make(map[uint]*cell),
		edges:    make(map[cell][]*cell),
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

			val, err := f.GetCellValue("Sheet1", axis)
			if err != nil {
				return nil, err
			}

			c := &cell{
				id:  concat(uint(colx), uint(rowx)),
				y:   uint(rowx),
				val: val,
			}

			graph.addCell(c, formula)
		}
		colx = 0
	}

	return nil, nil
}

func (d *digraph) addCell(c *cell, formula string) error {
	if formula != "" {
		// add cell.
		d.formulas[c.id] = c

		axis := digestFormula(formula)

		// TODO: attaching algorithm
		for _, pos := range axis {
			x, y, err := excelize.CellNameToCoordinates(pos)
			if err != nil {
				return err
			}

		}
	}
}

func digestFormula(formula string) []string {
	s := strings.FieldsFunc(formula, func(r rune) bool {
		i := int(r)
		if i < len(validRunes) {
			return validRunes[i]
		}
		return false
	})

	// iterate backwards to not worry about index shifting.
	for i := len(s) - 1; i >= 0; i-- {
		if s[i][0] >= 48 && s[i][0] <= 57 { // check if byte is number.
			s = append(s[:i], s[i+1:]...)
		}
	}

	return s
}

func concat(x, y uint) uint {
	var mul uint = 10
	for y >= mul {
		mul *= 10
	}
	return x*mul + y
}

func unConcat(xy, y uint) uint {
	var mul uint = 10
	for y >= mul {
		mul *= 10
	}

	dif := xy - y

	return dif / mul
}
