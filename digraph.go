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
	id       uint
	y        uint
	isCyclic bool
}

type digraph struct {
	f         *excelize.File
	formulas  map[uint]*cell
	relations map[*cell][]*cell
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
		f:         f,
		formulas:  make(map[uint]*cell),
		relations: make(map[*cell][]*cell),
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

			if formula == "" {
				continue
			}

			c := &cell{
				id: concat(uint(colx), uint(rowx)),
				y:  uint(rowx),
			}

			graph.addCell(c, formula)
		}
		colx = 0
	}

	return graph, nil
}

func (d *digraph) addCell(c *cell, formula string) error {
	if ptr, ok := d.formulas[c.id]; !ok {
		d.formulas[c.id] = c
	} else {
		c = ptr
	}
	references := digestFormula(formula)

	for _, refAxis := range references {
		formula, err := d.f.GetCellFormula("Sheet1", refAxis)
		if err != nil {
			return err
		}

		if formula == "" { // non-formula cells point to nothing so they cant cause a cycle.
			continue
		}

		x, y, err := excelize.CellNameToCoordinates(refAxis)
		if err != nil {
			return err
		}

		refID := concat(uint(x), uint(y))
		refCell, ok := d.formulas[refID]
		if !ok {
			refCell = &cell{
				id: refID,
				y:  uint(y),
			}
			d.formulas[refID] = refCell
		}

		d.relations[c] = append(d.relations[c], refCell)
	}

	return nil
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
