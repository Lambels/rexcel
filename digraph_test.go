package main

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestLoadRelations(t *testing.T) {
	if err := filepath.Walk("./testdata", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".xlsx" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		g, err := newGraph(f)
		if err != nil {
			return err
		}
		g.scc()

		// load results file.
		ext := filepath.Ext(path)
		resPath := strings.TrimSuffix(path, ext) + "_res.txt"
		res := loadResFile(t, resPath)

		for _, v := range res {
			x, y, err := excelize.CellNameToCoordinates(v)
			if err != nil {
				return err
			}
			id := concat(uint(x), uint(y))

			if _, ok := g.circular[id]; !ok {
				t.Fatalf("Expected to find %s in result set of %s", v, path)
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func loadResFile(t *testing.T, path string) []string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	return strings.Split(string(data), ",")
}
