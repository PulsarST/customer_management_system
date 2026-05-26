package models

type Order struct {
	Name     string  `db:name`
	Address  string  `db:address`
	Quantity int     `db:quantity`
	Cost     float32 `db:cost`
}
