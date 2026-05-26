package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func ConnectMockup() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "../internal/handlers/db/mockup.db")
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

// CREATING TABLES AND MOCKUP DATA
func GenerateMockUpData() {

	db := ConnectMockup()

	_, exec_err := db.Exec(`
		INSERT INTO Products (product_name, description) VALUES
		('Ноутбук Lenovo ThinkPad E14', '14-дюймовый бизнес ноутбук'),
		('Монитор Samsung 24"', 'IPS монитор 24 дюйма'),
		('Клавиатура Logitech K380', 'Беспроводная клавиатура'),
		('Мышь Logitech MX Master 3', 'Эргономичная беспроводная мышь'),
		('SSD Samsung 970 EVO 1TB', 'NVMe SSD накопитель'),
		('HDD Seagate 2TB', 'Жесткий диск 2TB'),
		('Роутер TP-Link Archer C6', 'Wi-Fi роутер'),
		('Принтер HP LaserJet M404', 'Лазерный принтер'),
		('Сканер Canon LiDE 300', 'Планшетный сканер'),
		('Веб-камера Logitech C920', 'FullHD веб-камера'),
		('Наушники Sony WH-1000XM4', 'Беспроводные наушники'),
		('Колонки Edifier R1280T', 'Полочная акустика'),
		('Смартфон Samsung Galaxy A54', 'Android смартфон'),
		('Планшет iPad Air', 'Планшет Apple'),
		('USB Flash Kingston 64GB', 'USB 3.2 флешка'),
		('Кабель HDMI 2м', 'HDMI кабель версии 2.1'),
		('Сетевой фильтр Defender', 'Фильтр питания на 5 розеток'),
		('Кресло офисное Chairman 405', 'Офисное кресло'),
		('Стол офисный IKEA Bekant', 'Офисный стол'),
		('Микрофон HyperX QuadCast', 'USB микрофон');

		INSERT INTO Storages (address) VALUES
		('г. Алматы, ул. Сатпаева 101'),
		('г. Алматы, пр. Абая 55'),
		('г. Астана, ул. Кабанбай Батыра 12'),
		('г. Шымкент, ул. Тауке Хана 77'),
		('г. Караганда, ул. Бухар Жырау 44');

		INSERT INTO Stock (product_id, storage_id, quantity) VALUES
		(1, 1, 12),
		(2, 1, 25),
		(3, 1, 40),
		(4, 2, 18),
		(5, 2, 30),
		(6, 2, 14),
		(7, 3, 20),
		(8, 3, 8),
		(9, 3, 11),
		(10, 4, 16),
		(11, 4, 9),
		(12, 4, 13),
		(13, 5, 22),
		(14, 5, 7),
		(15, 5, 50);
	`)
	if exec_err != nil {
		log.Printf("Query failed: %v", exec_err)
		log.Fatalln(exec_err)
	}

	db.Close()
}
