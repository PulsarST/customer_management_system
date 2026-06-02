package models

type Product struct {
	Id       int     `db:"product_id"`
	Name     string  `db:"product_name"`
	Category string  `db:"category"`
	Price    float32 `db:"price"`
}
