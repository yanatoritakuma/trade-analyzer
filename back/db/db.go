package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDB はPostgreSQL（ローカルDocker または Neon）へ接続したGORMインスタンスを返す。
// 接続先は DATABASE_URL を優先し、未設定の場合は NEON_DATABASE_URL を使用する。
func NewDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("NEON_DATABASE_URL")
	}
	if dsn == "" {
		log.Fatalln("DATABASE_URL（または NEON_DATABASE_URL）が設定されていません")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Connected to DB")
	return db
}

// CloseDB はGORMが保持するコネクションを閉じる。
func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalln(err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatalln(err)
	}
}
