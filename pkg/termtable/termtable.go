package termtable

type TermTable struct {
	Header []string
	Rows   [][]string
}

func New(headerNames ...string) *TermTable {
	return &TermTable{
		Header: headerNames,
	}
}

func (t *TermTable) AddRow(row ...string) {
	t.Rows = append(t.Rows, row)
}

func (t *TermTable) Print() {
	// Find the max length of each column
	colWidths := make([]int, len(t.Header))
	for i, header := range t.Header {
		colWidths[i] = len(header)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print the header
	for i, header := range t.Header {
		if i != 0 {
			print("| ")
		}
		print(header)
		print(" ")
		for j := 0; j < colWidths[i]-len(header); j++ {
			print(" ")
		}
	}
	print("\n")

	// Print the separator
	sumColWidths := (len(colWidths) - 1) * 3 // this 3 includes "| " and the " " after the cell
	for _, colWidth := range colWidths {
		sumColWidths += colWidth
	}

	for i := 0; i < sumColWidths; i++ {
		print("=")
	}
	print("\n")

	// Print the rows
	for _, row := range t.Rows {
		for i, cell := range row {
			if i != 0 {
				print("| ")
			}
			print(cell)
			print(" ")
			for j := 0; j < colWidths[i]-len(cell); j++ {
				print(" ")
			}
		}
		println()
	}
}
