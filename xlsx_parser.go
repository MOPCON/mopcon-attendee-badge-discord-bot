package main

import (
	"fmt"

	"github.com/tealeg/xlsx/v3"
)

type Ticket struct {
	Code  string
	Email string
	Badge string
}

var KeyPairMap map[string]Ticket = make(map[string]Ticket)

func ReadXLSXToTicketMap(filename string, keypairmap *map[string]Ticket) {
	wb, err := xlsx.OpenFile(filename)
	if err != nil {
		panic(err)
	}

	fmt.Println("Reading XLSX File: ", filename)
	sheet := wb.Sheets[0]

	sheet.ForEachRow(func(r *xlsx.Row) error {
		(*keypairmap)[r.GetCell(0).String()] = Ticket{
			Code:  r.GetCell(0).String(),
			Email: r.GetCell(1).String(),
			Badge: r.GetCell(2).String(),
		}
		return nil
	})
}
