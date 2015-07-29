package main

import (
	"fmt"
	"strings"
)

type Table struct {
	rows []Row
}

type Row struct {
	cells  []Cell
	header bool
}

type Cell struct {
	text string
}

func NewTable() *Table {
	return &Table{
		rows: make([]Row, 0),
	}
}

func (t *Table) Done() {
	cols := 0
	for _, row := range t.rows {
		if cols < len(row.cells) {
			cols = len(row.cells)
		}
	}

	widths := make([]int, cols)
	for _, row := range t.rows {
		for col, cell := range row.cells {
			if widths[col] < len(cell.text) {
				widths[col] = len(cell.text)
			}
		}
	}

	for _, row := range t.rows {
		for col, cell := range row.cells {
			if row.header {
				fmt.Printf("%s%s ", strings.Repeat("_", widths[col]-len(cell.text)), strings.ToUpper(cell.text))
			} else {
				fmt.Printf("%s%s ", cell.text, strings.Repeat(" ", widths[col]-len(cell.text)))
			}
		}
		fmt.Println()
	}
}

func (t *Table) AddRow() *Row {
	t.rows = append(t.rows, Row{
		cells: make([]Cell, 0),
	})
	return &t.rows[len(t.rows)-1]
}

func (t *Table) AddRowWith(labels ...string) *Row {
	row := t.AddRow()
	for _, label := range labels {
		row.AddCellWithf("%s", label)
	}
	return row
}

func (r *Row) Header() {
	r.header = true
}

func (r *Row) AddCell() *Cell {
	r.cells = append(r.cells, Cell{
		text: "-",
	})
	return &r.cells[len(r.cells)-1]
}

func (r *Row) AddCellWithf(format string, a ...interface{}) *Cell {
	cell := r.AddCell()
	cell.Printf(format, a...)
	return cell
}

func (c *Cell) Printf(format string, a ...interface{}) {
	c.text = fmt.Sprintf(format, a...)
}
