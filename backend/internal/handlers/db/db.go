package db

import (
	"customer_managment_system/internal/models"
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func ConnectMockup() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "internal/handlers/db/mockup.db")
	if err != nil {
		log.Fatalln("cannot prodide sql conntection: ", err)
	}
	return db
}

// CREATING TABLES AND MOCKUP DATA
func GenerateMockUpData() {

	db := ConnectMockup()

	_, exec_err := db.Exec(`
		INSERT INTO Storages (storage_id, address) VALUES
(1, 'ул. Ленина, д. 10 (Главный склад)'),
(2, 'ул. Пушкина, д. 45 (Склад филиала Центр)'),
(3, 'пр. Просвещения, д. 12 (Северный терминал)'),
(4, 'ул. Гагарина, д. 88 (Южный ангар)'),
(5, 'ул. Чехова, д. 14 (Пункт выдачи и склад)'),
(6, 'ул. Кирова, д. 3 (Западный склад)'),
(7, 'пр. Строителей, д. 21 (Склад крупного опта)'),
(8, 'ул. Новая, д. 5 (Дополнительный склад)'),
(9, 'ул. фабричная, д. 12/2 (Склад при производстве)'),
(10, 'ул. Советская, д. 50 (Резервный фонд)');

-- ==========================================
-- 3. ЗАПОЛНЕНИЕ ТАБЛИЦЫ ТОВАРОВ (150 товаров)
-- ==========================================

WITH RECURSIVE
cnt(x) AS (
   SELECT 1
   UNION ALL
   SELECT x+1 FROM cnt WHERE x < 150
),
categories(id, name) AS (
   VALUES 
   (0, 'Письменные принадлежности'),
   (1, 'Бумажная продукция'),
   (2, 'Офисная техника и девайсы'),
   (3, 'Папки и системы архивации'),
   (4, 'Настольные аксессуары')
),
names(id, name, cat_id, price) AS (
   VALUES 
   (1, 'Ручка шариковая синяя', 0, 45.0), (2, 'Карандаш графитный HB', 0, 25.0), 
   (3, 'Маркер перманентный черный', 0, 85.0), (4, 'Тетрадь в клетку 48л', 1, 60.0),
   (5, 'Блокнот А5 на спирали', 1, 150.0), (6, 'Бумага А4 SvetoCopy (500л)', 1, 450.0),
   (7, 'Степлер до 20 листов', 2, 290.0), (8, 'Дырокол металлический', 2, 350.0),
   (9, 'Калькулятор инженерный', 2, 850.0), (10, 'Папка-регистратор 75мм', 3, 220.0),
   (11, 'Папка с файлами 40 шт', 3, 130.0), (12, 'Лоток для бумаг вертикальный', 4, 310.0),
   (13, 'Органайзер настольный', 4, 420.0), (14, 'Ножницы офисные 18см', 4, 110.0),
   (15, 'Клей-карандаш 21г', 4, 55.0)
)
INSERT INTO Products (product_id, product_name, category, description, price)
SELECT 
    cnt.x AS product_id,
    n.name || ' (Модель ' || cnt.x || ')' AS product_name,
    c.name AS category,
    'Качественный товар для офиса и школы. Артикул №' || (1000 + cnt.x) AS description,
    ROUND(n.price * (0.8 + (cnt.x % 5) * 0.1), 2) AS price -- Немного варьируем цены
FROM cnt
JOIN names n ON (cnt.x % 15) + 1 = n.id
JOIN categories c ON n.cat_id = c.id;

-- ==========================================
-- 4. ЗАПОЛНЕНИЕ ТАБЛИЦЫ ОСТАТКОВ (Stock)
-- ==========================================

-- Шаг A: Распределяем первые 130 товаров по складам 
-- (20 товаров с ID от 131 до 150 останутся без записей в Stock, т.е. нигде не хранятся)

-- Делаем так, чтобы часть товаров хранилась сразу в нескольких местах (условие задачи)
INSERT INTO Stock (product_id, storage_id, quantity)
SELECT 
    p.product_id,
    s.storage_id,
    (abs(random() % 100) + 5) AS quantity -- Случайное количество от 5 до 104
FROM Products p
CROSS JOIN Storages s
WHERE p.product_id <= 130  -- Исключаем последние 20 товаров
  AND (
    -- Условие, чтобы распределить товары по складам неравномерно:
    -- Товар 1-40 будет лежать на складах 1 и 2 (в двух хранилищах)
    (p.product_id BETWEEN 1 AND 40 AND s.storage_id IN (1, 2)) OR
    
    -- Товар 41-70 будет лежать на складах 3, 4 и 5 (в трех хранилищах)
    (p.product_id BETWEEN 41 AND 70 AND s.storage_id IN (3, 4, 5)) OR
    
    -- Товар 71-100 будет лежать только на складе 6 (в одном хранилище)
    (p.product_id BETWEEN 71 AND 100 AND s.storage_id = 6) OR
    
    -- Товар 101-130 разбросаем по остальным складам (7, 8, 9, 10) на основе остатка от деления
    (p.product_id BETWEEN 101 AND 130 AND s.storage_id = 7 + (p.product_id % 4))
  );
	`)
	if exec_err != nil {
		log.Printf("Query failed: %v", exec_err)
	}

	db.Close()
}

func LoadOrderToDb(orders []models.Order) {
	db := ConnectMockup()

	_, err := db.NamedExec(`INSERT INTO Orders (order_date, product_name, storage_address, quantity, cost) 
		VALUES (CURRENT_TIMESTAMP, :product_name, :storage_address, :quantity, :cost)`, orders)

	if err != nil {
		log.Println(err.Error())
	}

	db.Close()
}

func GetProducts() []models.Product {
	db := ConnectMockup()

	var result []models.Product

	err := db.Select(&result, `SELECT * FROM Products`)

	if err != nil {
		log.Println(err.Error())
	}

	db.Close()

	return result
}

func GetStorages() []models.Storage {
	db := ConnectMockup()

	var result []models.Storage

	err := db.Select(&result, `SELECT * FROM Storages`)

	if err != nil {
		log.Println(err.Error())
	}

	db.Close()

	return result
}

func GetStocks() []models.Stock {
	db := ConnectMockup()

	var result []models.Stock

	err := db.Select(&result, `SELECT * FROM Stock`)

	if err != nil {
		log.Println(err.Error())
	}

	db.Close()

	return result
}

func GetSchemaString() string {
	products, _ := json.Marshal(GetProducts())
	storages, _ := json.Marshal(GetStorages())
	stocks, _ := json.Marshal(GetStocks())

	return `Products: ` + string(products) + `\n Storages: ` + string(storages) + `\n Stock: ` + string(stocks)
}
