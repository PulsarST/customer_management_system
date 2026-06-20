package models

type Order struct {
	Id       int     `db:"order_id"`
	Name     string  `db:"product_name"`
	Address  string  `db:"storage_address"`
	Quantity int     `db:"quantity"`
	Cost     float32 `db:"cost"`
}
