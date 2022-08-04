package log

import "fmt"

type Pos struct {
	Line int
	Column int
}


type Diagnostic struct {
	Error error
	Description string 
	Pos Pos
}

type Diagnostics []Diagnostic

func (d Diagnostics) Len() int {
	return len(d)
}

func (d Diagnostics) HasError() bool {
	for _, diag := range d {
		if diag.Error != nil {
			return true
		}
	}
	return false
}


func (d Diagnostics) Show() {
	for _, diag := range d {
		if diag.Error != nil {
			fmt.Printf("%s\n", diag.Error)
		}
	}
}
