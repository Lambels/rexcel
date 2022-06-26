package main

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/xuri/excelize/v2"
)

// validRunes runes represents a probably incomplete maping of runes to be ommited inside formulas.
var validRunes = [...]bool{
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'(': true,
	')': true,
}

// cell represents a formula excel cell.
type cell struct {
	// id is formed by cocatonating the x and y of the cell.
	// used to calculate lowlink value.
	id uint
	// y is used to obtain the x from the id.
	y uint
	// isCyclic indicates if the formula cell ultimately leads in a
	// recursive function.
	isCyclic bool

	lowlink uint
	onStack bool
}

// String representation for cell.
func (c *cell) String() string {
	axis, _ := excelize.CoordinatesToCellName(int(unConcat(c.id, c.y)), int(c.y))
	return fmt.Sprintf(
		"%s cyclic: %t onStack: %t lowLink %d id %d",
		axis,
		c.isCyclic,
		c.onStack,
		c.lowlink,
		c.id,
	)
}

// digraph represents a directed graph of formula cells inside an excel file.
type digraph struct {
	f         *excelize.File
	formulas  map[uint]*cell
	relations map[*cell][]*cell
	circular  []*cell
	stack     []*cell
}

// newGraph forms a graph relationship for each of the formula cells in the reader.
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
		stack:     make([]*cell, 0),
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

			id := concat(uint(colx), uint(rowx))
			c := &cell{
				id:      id,
				y:       uint(rowx),
				lowlink: id,
			}

			graph.addCell(c, formula)
		}
		colx = 0
	}

	return graph, nil
}

// addCell adds a formula cell to the graph whilst keeping pointers consistent throughout the
// graph realations.
// addCell will also add the referenced cells of the provided cell which are formula cells.
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
				id:      refID,
				y:       uint(y),
				lowlink: refID,
			}
			d.formulas[refID] = refCell
		}

		d.relations[c] = append(d.relations[c], refCell)
	}

	return nil
}

func (d *digraph) scc() {
	for c := range d.relations {
		visited := make(map[uint]bool)
		results := make([][]*cell, 0)
		d.dfs(c, visited, &results)

		for _, c := range d.stack {
			c.onStack = false
			c.lowlink = c.id
		}
		d.stack = d.stack[:0]
	}
}

func (d *digraph) dfs(node *cell, visited map[uint]bool, results *[][]*cell) {
	visited[node.id] = true
	d.stack = append(d.stack, node)
	node.onStack = true

	for _, c := range d.relations[node] {
		if !visited[c.id] {
			d.dfs(c, visited, results)
		} else if c.onStack {
			node.lowlink = uint(math.Min(float64(c.lowlink), float64(node.lowlink)))
		}
	}

	if node.lowlink == node.id {
		i := len(d.stack) - 1
		var vertices []*cell
		for {
			n := d.stack[i]
			n.onStack = false
			n.lowlink = n.id
			d.stack = d.stack[:i]
			vertices = append(vertices, n)
			if node.id == n.id {
				break
			}
			i--
		}

		*results = append(*results, vertices)
	}
}

// digestFormula digests the formula: formula to get the referenced cells in the formula.
func digestFormula(formula string) []string {
	// filter out opperations.
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

// concat takes x, y uint s and produces an uint of the form xy.
// ex: x = 123 , y = 456 -> xy = 123456
func concat(x, y uint) uint {
	var mul uint = 10
	for y >= mul {
		mul *= 10
	}
	return x*mul + y
}

// unConcat takes xy, y uint s and undo s the concatonation process.
// given the concatonated xy we can take away y to remain with x and y.
// ex: xy = 123456 , y = 456 -> x = 123
func unConcat(xy, y uint) uint {
	var mul uint = 10
	for y >= mul {
		mul *= 10
	}

	dif := xy - y

	return dif / mul
}
