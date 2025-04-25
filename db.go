package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	//"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB //Переменная сессии БД

const (
	host     = "localhost"
	port     = 5432
	user     = "admin"
	password = "12345678"
	dbname   = "meowdb"
)

func ConnectDB() error {
	var err error
	db, err = sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname))
	fmt.Println("Successfully connected!")
	return err
}

func generateSalt() string {
	salt := make([]byte, 16) // 16 байт соли
	_, err := rand.Read(salt)
	if err != nil {
		log.Fatal("Ошибка генерации соли:", err)
	}
	return hex.EncodeToString(salt)
}

// Функция для проверки введенного ключа
func verifyKey(userKey string) (int, bool, error) {
	// 1. Извлекаем все записи с ключами и солями из базы данных
	rows, err := db.Query(`SELECT id_key, content_key, salt FROM key`)
	if err != nil {
		return 0, false, fmt.Errorf("ошибка при извлечении ключей из базы данных: %w", err)
	}
	defer rows.Close()
	fmt.Println("Записи считаны")
	// 2. Перебираем все записи в базе данных
	for rows.Next() {
		var keyID int
		var storedHashedKey, storedSalt string
		if err := rows.Scan(&keyID, &storedHashedKey, &storedSalt); err != nil {
			return 0, false, fmt.Errorf("ошибка при чтении данных из базы: %w", err)
		}
		fmt.Println(storedHashedKey)
		//fmt.Println(storedSalt)
		// 3. Захешируем введенный ключ с этой солью
		h := sha256.New()
		h.Write([]byte(userKey + storedSalt))
		hashedInputKey := hex.EncodeToString(h.Sum(nil))
		fmt.Println(hashedInputKey)
		// 4. Сравниваем полученный хеш с тем, что хранится в базе данных
		if storedHashedKey == hashedInputKey {
			fmt.Println("Ключ совпал блин!!1")
			return keyID, true, nil // ключ совпал
		}
	}
	// 5. Если ни один ключ не совпал, возвращаем false
	return 0, false, nil
}

// Функция добавления аккаунта в базу данных
func addAccountToDB(login, password, contentKey string) error {

	// Проверяем, есть ли уже такой логин
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM account WHERE login=$1)", login)
	if err != nil {
		return fmt.Errorf("ошибка проверки логина: %w", err)
	}
	if exists {
		return fmt.Errorf("логин уже существует")
	}

	// Проверяем ключ и получаем keyID
	keyID, isValid, err := verifyKey(contentKey)
	if err != nil {
		return fmt.Errorf("ошибка при проверке ключа: %w", err)
	}
	if !isValid {
		return fmt.Errorf("неверный ключ")
	}
	salt := generateSalt()
	// Хеширование пароля с использованием SHA-256
	h := sha256.New()
	h.Write([]byte(password + salt))
	hashedPassword := hex.EncodeToString(h.Sum(nil))

	// Добавляем пользователя
	query := "INSERT INTO account (login, password, salt, id_key) VALUES (($1), ($2), ($3), ($4))"
	_, err = db.Exec(query, login, hashedPassword, salt, keyID)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении аккаунта: %w", err)
	}

	fmt.Println("Account added in table: ", login)
	return nil
}

// Функция для проверки введенного логина и пароля
func verifyLoginCredentials(login, password string) error {
	var storedSalt, storedHashedPassword string
	err := db.QueryRow("SELECT salt, password FROM account WHERE login = $1", login).Scan(&storedSalt, &storedHashedPassword)
	if err == sql.ErrNoRows {
		return fmt.Errorf("пользователь с таким логином не найден")
	} else if err != nil {
		return fmt.Errorf("ошибка при извлечении данных из базы: %w", err)
	}

	// Хешируем введенный пароль с использованием той же соли
	h := sha256.New()
	h.Write([]byte(password + storedSalt))
	hashedInputPassword := hex.EncodeToString(h.Sum(nil))

	// Сравниваем полученный хеш с тем, который хранится в базе данных
	if storedHashedPassword != hashedInputPassword {
		return fmt.Errorf("неверный пароль")
	}

	return nil
}

// Определение структуры ссылки на игру в меню
type GameRef struct {
	Name string `json:"Name"`
	Icon string `json:"Icon"`
	Type string `json:"Type"`
	Path string `json:"Path"`
}

// Определение структур JSON
type Answer struct {
	IsRight bool   `json:"IsRight,omitempty"`
	Image   string `json:"Image,omitempty"`
	Phrase  string `json:"Phrase"`
	Name    string `json:"Name,omitempty"`
}

type Answers []Answer

func (a *Answers) UnmarshalJSON(data []byte) error {
	// Попробуем сначала распарсить как массив
	var multiple []Answer
	if err := json.Unmarshal(data, &multiple); err == nil {
		*a = multiple
		return nil
	}

	// Если не удалось, пробуем как объект
	var single Answer
	if err := json.Unmarshal(data, &single); err == nil {
		*a = []Answer{single}
		return nil
	}

	return fmt.Errorf("не удалось разобрать Answers: неверный формат JSON")
}

type Page struct {
	MainImage  string  `json:"MainImage"`
	MainPhrase string  `json:"MainPhrase"`
	Answers    Answers `json:"Answers"`
}

type Game struct {
	Name  string `json:"Name"`
	Icon  string `json:"Icon,omitempty"`
	Type  string `json:"Type"`
	Pages []Page `json:"Pages"`
}

var gamePathByName = make(map[string]string)

// Подключение к базе данных

// Обход всех JSON-файлов в папке
func processAllJsonFiles(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}
		fmt.Println("Обрабатываю файл:", path)
		processJsonFile(path)
		return nil
	})
}

// Обработка одного JSON-файла
func processJsonFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Ошибка открытия файла:", err)
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Ошибка чтения файла:", err)
		return
	}

	// Определяем тип JSON-файла (игра или категория)
	var temp struct {
		Type string `json:"Type"`
	}
	if err := json.Unmarshal(bytes, &temp); err != nil {
		log.Println("Ошибка определения типа JSON:", err)
		return
	}

	if temp.Type == "SubCategory" {
		processMenuJson(filePath, bytes)
	} else {
		processGameJson(filePath, bytes)
	}
}

// Обработка JSON-файла категории (меню)
func processMenuJson(filePath string, bytes []byte) {
	var category struct {
		Name      string `json:"Name"`
		MainImage string `json:"MainImage"`
		Pages     []struct {
			TopRight    *GameRef `json:"TopRight"`
			CenterRight *GameRef `json:"CenterRight"`
			BottomRight *GameRef `json:"BottomRight"`
			TopLeft     *GameRef `json:"TopLeft"`
			CenterLeft  *GameRef `json:"CenterLeft"`
			BottomLeft  *GameRef `json:"BottomLeft"`
		} `json:"Pages"`
	}

	if err := json.Unmarshal(bytes, &category); err != nil {
		log.Println("Ошибка декодирования JSON категории:", err)
		return
	}

	// Добавляем категорию в БД
	var categoryID int
	err := db.QueryRow("INSERT INTO category (tag, icon) VALUES ($1, $2) ON CONFLICT (tag) DO UPDATE SET icon=EXCLUDED.icon RETURNING id_category",
		category.Name, category.MainImage).Scan(&categoryID)
	if err != nil {
		log.Println("Ошибка добавления категории:", err)
		return
	}

	fmt.Println("Категория добавлена:", category.Name)

	// Обрабатываем вложенные игры
	gameRefs := []*GameRef{}
	for _, page := range category.Pages {
		gameRefs = append(gameRefs, page.TopRight, page.CenterRight, page.BottomRight, page.TopLeft, page.CenterLeft, page.BottomLeft)
	}

	for _, gameRef := range gameRefs {
		if gameRef != nil && gameRef.Path != "" {
			gamePath := filepath.Join(filepath.Dir(filePath), gameRef.Path)
			gamePath = filepath.Clean(gamePath) // Очистка пути
			processJsonFile(gamePath)           // Рекурсивно загружаем игру
		}
	}
}

// Обработка JSON-файла игры
func processGameJson(filePath string, bytes []byte) {
	var game Game
	if err := json.Unmarshal(bytes, &game); err != nil {
		log.Println("Ошибка декодирования JSON игры:", err)
		return
	}

	// Определяем категории по пути
	categoryNames := getCategoryNamesFromPath(filePath)

	// Добавляем игру в БД
	var gameID int
	err := db.QueryRow("INSERT INTO games (name_game, type, icon, json_path) VALUES ($1, $2, $3, $4) ON CONFLICT (name_game) DO UPDATE SET name_game=EXCLUDED.name_game RETURNING id_game",
		game.Name, game.Type, game.Icon, filePath).Scan(&gameID)
	if err != nil {
		log.Println("Ошибка добавления игры:", err)
		return
	}

	// Добавляем все категории, связанные с игрой
	for _, categoryName := range categoryNames {
		var categoryID int
		err := db.QueryRow("INSERT INTO category (tag) VALUES ($1) ON CONFLICT (tag) DO UPDATE SET tag=EXCLUDED.tag RETURNING id_category", categoryName).Scan(&categoryID)
		if err != nil {
			log.Println("Ошибка добавления категории:", err)
			continue
		}

		// Связываем игру с категорией
		_, err = db.Exec("INSERT INTO game_category (id_game, id_category) VALUES ($1, $2) ON CONFLICT DO NOTHING", gameID, categoryID)
		if err != nil {
			log.Println("Ошибка связывания игры с категорией:", err)
		}
	}

	// Добавляем изображения
	addImage(gameID, game.Icon)
	for _, page := range game.Pages {
		addImage(gameID, page.MainImage)
		for _, answer := range page.Answers {
			addImage(gameID, answer.Image)
		}
	}

	fmt.Println("Успешно добавлена игра:", game.Name)
}

// Функция добавления изображения в БД
func addImage(gameID int, imageName string) {
	if imageName == "" {
		return
	}
	_, err := db.Exec("INSERT INTO images (id_game, image_name) VALUES ($1, $2) ON CONFLICT DO NOTHING", gameID, imageName)
	if err != nil {
		log.Println("Ошибка добавления изображения:", err)
	}
}

// Получаем категории из пути файла
func getCategoryNamesFromPath(filePath string) []string {
	dir := filepath.Dir(filePath)
	return strings.Split(filepath.Base(dir), ", ") // Если категории разделены запятой
}

//
//
//
//
//

//
//
//
//
//

//
//
//
//
//

//
//
//
//
//

//
//
//
//
//

//
//
//
//
//

//
//
//
//
//

/*
// Обход всех JSON-файлов в папке
func processAllJsonFiles(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".json") {
			fmt.Println("Обрабатываю файл:", path)
			processJsonFile(path)
		}
		return nil
	})
}

// Обработка одного JSON-файла
func processJsonFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Ошибка открытия файла:", err)
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Ошибка чтения файла:", err)
		return
	}

	var game Game
	if err := json.Unmarshal(bytes, &game); err != nil {
		log.Println("Ошибка декодирования JSON:", err)
		return
	}

	// Определяем категорию по пути
	categoryName := filepath.Base(filepath.Dir(filePath))

	// Добавляем категорию в БД, если её нет
	var categoryID int
	err = db.QueryRow("INSERT INTO category (tag) VALUES ($1) ON CONFLICT (tag) DO UPDATE SET tag=EXCLUDED.tag RETURNING id_category", categoryName).Scan(&categoryID)
	if err != nil {
		log.Println("Ошибка добавления категории:", err)
		return
	}

	// Добавляем игру в БД
	var gameID int
	err = db.QueryRow("INSERT INTO games (category_id, name_game, type, icon, json_path)VALUES ($1, $2, $3, $4, $5) ON CONFLICT (name_game) DO UPDATE SET name_game=EXCLUDED.name_game RETURNING category_id",
		categoryID, game.Name, game.Type, game.Icon, filePath).Scan(&gameID)
	if err != nil {
		log.Println("Ошибка добавления игры:", err)
		return
	}

	// Добавляем изображения
	for _, page := range game.Pages {
		for _, answer := range page.Answers {
			//imagePath := filepath.Join(filepath.Base(filepath.Dir(filePath)), answer.Image)
			_, err := db.Exec("INSERT INTO images (id_game, image_name) VALUES ($1, $2)", gameID, answer.Image)
			if err != nil {
				log.Println("Ошибка добавления изображения:", err)
			}
		}
	}

	fmt.Println("Успешно добавлена игра:", game.Name)
}
*/

// Функция для загрузки картинок в БД
/*func uploadImagesToDB(dir string) error {
	insertStmt := "INSERT INTO images (id, name, path) VALUES (DEFAULT, $1, $2)"

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			//ext := filepath.Ext(info.Name()) // Получаем расширение файла
			/*if ext != ".png" {               // Оставляем только PNG
				return nil
			}*

			_, err := db.Exec(insertStmt, info.Name(), path)
			if err != nil {
				fmt.Println("Ошибка при добавлении в БД:", err)
			} else {
				fmt.Println("Добавлено:", info.Name())
			}
		}
		return nil
	})
}*/

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
/*// Функция генерации ключа
func generateKey() string {
	keyBytes := make([]byte, 16) // Генерация 16-байтового ключа
	_, err := rand.Read(keyBytes)
	if err != nil {
		log.Fatal("Ошибка генерации ключа:", err)
	}
	return hex.EncodeToString(keyBytes)
}

// Функция для добавления ключей в базу данных
func addKeysToDB(numKeys int) error {
	for i := 0; i < numKeys; i++ {
		// Генерируем новый ключ
		contentKey := generateKey()

		// Генерируем соль
		salt := generateSalt()

		// Хешируем ключ с солью
		h := sha256.New()
		h.Write([]byte(contentKey + salt))
		hashedKey := hex.EncodeToString(h.Sum(nil))

		// Генерация случайного типа доступа (можно заменить на другие значения)
		accessType := i % 3 // например, для демонстрации, случайно выбираем 0, 1 или 2

		// Добавляем ключ в таблицу key
		query := "INSERT INTO key (content_key, salt, access_type) VALUES ($1, $2, $3)"
		_, err := db.Exec(query, hashedKey, salt, accessType)
		if err != nil {
			return fmt.Errorf("ошибка при добавлении ключа: %w", err)
		}

		// Выводим информацию о сгенерированном ключе
		fmt.Printf("Ключ %d: %s (access_type=%d)\n", i+1, contentKey, accessType)
	}
	return nil
}*/
