package database

import (
	"log"
    "fmt"
	"x-ui/config"
	"x-ui/database/model"
	"x-ui/xray"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

const (
	defaultUsername = "admin"
	defaultPassword = "admin"
	defaultSecret   = ""
)

func initModels() error {
	models := []interface{}{
		&model.User{},
		&model.Inbound{},
		&model.OutboundTraffics{},
		&model.Setting{},
		&model.InboundClientIps{},
		&xray.ClientTraffic{},
	}
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Error auto migrating model: %v", err)
			return err
		}
	}
	return nil
}

func initUser() error {
	empty, err := isTableEmpty("users")
	if err != nil {
		log.Printf("Error checking if users table is empty: %v", err)
		return err
	}
	if empty {
		user := &model.User{
			Username:    defaultUsername,
			Password:    defaultPassword,
			LoginSecret: defaultSecret,
		}
		return db.Create(user).Error
	}
	return nil
}

func isTableEmpty(tableName string) (bool, error) {
	var count int64
	err := db.Table(tableName).Count(&count).Error
	return count == 0, err
}

func InitDB(dsn string) error {
	// Проверяем, что строка подключения не пуста
	if dsn == "" {
		return fmt.Errorf("empty database DSN")
	}

	// Логирование уровня отладки
	var gormLogger logger.Interface
	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	// Логируем строку подключения для отладки (осторожно, не выводить пароли в логах)
	log.Printf("Connecting to database with DSN: %s", dsn)

	// Создаем конфигурацию для GORM
	c := &gorm.Config{
		Logger: gormLogger,
	}

	// Открываем соединение с PostgreSQL
	var err error
	db, err = gorm.Open(postgres.Open(dsn), c)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Инициализация моделей и пользователя
	if err := initModels(); err != nil {
		return fmt.Errorf("failed to initialize models: %v", err)
	}

	if err := initUser(); err != nil {
		return fmt.Errorf("failed to initialize user: %v", err)
	}

	// Возвращаем успешный результат
	return nil
}


func CloseDB() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func GetDB() *gorm.DB {
	return db
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func Checkpoint() error {
	err := db.Exec("PRAGMA wal_checkpoint;").Error
	if err != nil {
		return err
	}
	return nil
}
