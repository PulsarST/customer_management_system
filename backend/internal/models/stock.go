package models

type Stock struct {
	ProductID int `db:"product_id"`
	StorageID int `db:"storage_id"`
	Quantity  int `db:"quantity"`
}
