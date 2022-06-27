package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

const usageMessage string = `USAGE:
rexcel <fileName ...> an arbitrary amound of valid .xlsx file names can be provided.
`

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		exitUsage(errors.New("error: not enough arguments were supplied"))
	}

	res, err := processFiles(args)
	if err != nil {
		exitUsage(err)
	}

	for _, r := range res {
		fmt.Println(r.fName)
		for _, cell := range r.results {
			fmt.Println(cell)
		}
	}
}

type result struct {
	fName   string
	results []string
}

func processFiles(paths []string) ([]result, error) {
	results := make([]result, 0)
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("Error while opening file %s: %w", path, err)
		}

		g, err := newGraph(f)
		if err != nil {
			return nil, fmt.Errorf("Error while processing file %s: %w", path, err)
		}
		g.scc()

		r := result{
			fName:   path,
			results: make([]string, 0),
		}

		for id := range g.circular {
			c := g.formulas[id]
			axis, err := excelize.CoordinatesToCellName(
				int(unConcat(c.id, c.y)),
				int(c.y),
			)
			if err != nil {
				return nil, err
			}

			r.results = append(r.results, axis)
		}

		results = append(results, r)
	}

	return results, nil
}

func exitUsage(err error) {
	fmt.Println(err)
	fmt.Println(usageMessage)
	os.Exit(1)
}
