package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const schema = `
	CREATE TABLE Products(
		product_id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
		product_name VARCHAR(200) NOT NULL,
		product_description TEXT
	);

	CREATE TABLE Orders(
		order_id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT
		product_name varchar(100) NOT NULL
		storage_address varchar(100) NOT NULL
		quantity int NOT NULL
		cost numeric NOT NULL
	);

	CREATE TABLE Storages(
		storage_id BIGINT NOT NULL PRIMARY KEY autoincrement
		address varchar(200) NOT NULL
	);

	CREATE TABLE Stock(
		product_id BIGINT NOT NULL REF
		storage_id BIGINT NOT NULL REF
		quantity int NOT NULL
	);
`

// CREATING TABLES AND MOCKUP DATA
func main() {
	db, err := sqlx.Connect("sqlite", "")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)
}
