package db

import (
	"customer_managment_system/internal/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	dbOnce     sync.Once
	dbInstance *sqlx.DB
	dbInitErr  error
)

// resolveDBPath locates mockup.db robustly: honours the DB_PATH env var, then
// walks up from both the executable directory and the working directory looking
// for the database. This makes the server runnable from any directory instead
// of only from backend/.
func resolveDBPath() string {
	if p := os.Getenv("DB_PATH"); p != "" {
		return p
	}

	const maxLevels = 6
	rel := filepath.Join("internal", "handlers", "db")
	const marker = "mockup.db"

	var starts []string
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}

	for _, start := range starts {
		dir := start
		for i := 0; i < maxLevels; i++ {
			// Either the full backend layout, or mockup.db sitting next to a dir.
			if _, err := os.Stat(filepath.Join(dir, rel, marker)); err == nil {
				return filepath.Join(dir, rel, marker)
			}
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return filepath.Join(dir, marker)
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	// Nothing found; return the conventional path so the error message is clear.
	return filepath.Join(rel, marker)
}

// ConnectMockup returns a process-wide shared *sqlx.DB. The connection is opened
// once (lazily) and reused; sqlx/database-sql is safe for concurrent use, so this
// avoids opening and closing a fresh handle on every request.
func ConnectMockup() (*sqlx.DB, error) {
	dbOnce.Do(func() {
		path := resolveDBPath()
		dbInstance, dbInitErr = sqlx.Connect("sqlite", path)
		if dbInitErr != nil {
			log.Printf("cannot open sqlite database at %q: %v", path, dbInitErr)
		}
	})
	return dbInstance, dbInitErr
}

// CREATING TABLES AND MOCKUP DATA
func GenerateMockUpData() {

	db, err := ConnectMockup()
	if err != nil {
		return
	}

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
}

func LoadOrderToDb(orders []models.Order) {
	if len(orders) == 0 {
		return
	}

	db, err := ConnectMockup()
	if err != nil {
		return
	}

	if _, err := db.NamedExec(`INSERT INTO Orders (order_date, product_name, storage_address, quantity, cost)
		VALUES (CURRENT_TIMESTAMP, :product_name, :storage_address, :quantity, :cost)`, orders); err != nil {
		log.Println(err.Error())
	}
}

func GetProducts() []models.Product {
	db, err := ConnectMockup()
	if err != nil {
		return nil
	}

	var result []models.Product

	if err := db.Select(&result, `SELECT * FROM Products`); err != nil {
		log.Println(err.Error())
	}

	return result
}

func GetStorages() []models.Storage {
	db, err := ConnectMockup()
	if err != nil {
		return nil
	}

	var result []models.Storage

	if err := db.Select(&result, `SELECT * FROM Storages`); err != nil {
		log.Println(err.Error())
	}

	return result
}

func GetStocks() []models.Stock {
	db, err := ConnectMockup()
	if err != nil {
		return nil
	}

	var result []models.Stock

	if err := db.Select(&result, `SELECT * FROM Stock`); err != nil {
		log.Println(err.Error())
	}

	return result
}

// GetSchemaString renders the catalog for the LLM prompt in a compact,
// token-efficient form. The previous implementation dumped full JSON (repeating
// every key and quote on all ~380 rows), which cost ~11k prompt tokens per
// message. A TSV layout plus a category legend conveys the same data in roughly
// half the tokens, which directly cuts prefill latency.
func GetSchemaString() string {
	products := GetProducts()
	storages := GetStorages()
	stocks := GetStocks()

	// Build a category legend so long category names are not repeated on every
	// product row; products reference the short category id instead.
	catIDs := map[string]int{}
	var catNames []string
	for _, p := range products {
		if _, ok := catIDs[p.Category]; !ok {
			catIDs[p.Category] = len(catNames)
			catNames = append(catNames, p.Category)
		}
	}

	var b strings.Builder

	b.WriteString("Categories (id=name):\n")
	for id, name := range catNames {
		fmt.Fprintf(&b, "%d=%s\n", id, name)
	}

	b.WriteString("\nProducts (TSV: product_id, category_id, price, product_name):\n")
	for _, p := range products {
		fmt.Fprintf(&b, "%d\t%d\t%g\t%s\n", p.Id, catIDs[p.Category], p.Price, p.Name)
	}

	b.WriteString("\nStorages (TSV: storage_id, address):\n")
	for _, s := range storages {
		fmt.Fprintf(&b, "%d\t%s\n", s.Id, s.Address)
	}

	// Collapse Stock into one line per product: which storages hold it. The exact
	// per-storage quantity is not needed for recommending products or resolving a
	// storage address, so omitting it saves a large number of tokens.
	storagesByProduct := map[int][]int{}
	for _, st := range stocks {
		storagesByProduct[st.ProductID] = append(storagesByProduct[st.ProductID], st.StorageID)
	}
	productIDs := make([]int, 0, len(storagesByProduct))
	for pid := range storagesByProduct {
		productIDs = append(productIDs, pid)
	}
	sort.Ints(productIDs)

	b.WriteString("\nStock (product_id -> storage_ids that have it in stock):\n")
	for _, pid := range productIDs {
		ids := storagesByProduct[pid]
		parts := make([]string, len(ids))
		for i, id := range ids {
			parts[i] = fmt.Sprintf("%d", id)
		}
		fmt.Fprintf(&b, "%d -> %s\n", pid, strings.Join(parts, ","))
	}

	return b.String()
}
