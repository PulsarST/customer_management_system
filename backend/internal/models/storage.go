package models

type Storage struct {
	Id      int    `db:"storage_id"`
	Address string `db:"address"`
}
