package models

type Product struct {
	Id          int     `db:"product_id"`
	Name        string  `db:"product_name"`
	Category    string  `db:"category"`
	Description string  `db:"description"`
	Price       float32 `db:"price"`
}
