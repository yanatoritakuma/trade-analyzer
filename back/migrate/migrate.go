package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/yanatoritakuma/trade-analyzer/back/db"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// AutoMigrate でテーブルを作成する。`go run migrate/migrate.go` で実行する。
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env を読み込めませんでした（環境変数を直接使用します）")
	}

	database := db.NewDB()
	defer db.CloseDB(database)

	if err := database.AutoMigrate(
		&model.User{},
		&model.InvitationCode{},
		&model.Watchlist{},
		&model.WatchlistCandidate{},
		&model.StockPrice{},
		&model.Trade{},
		&model.RealPosition{},
		&model.AnalysisLog{},
		&model.LearningLog{},
		&model.LearningVersion{},
		&model.AnalysisSetting{},
		&model.AnalysisTheme{},
	); err != nil {
		log.Fatalln(err)
	}

	log.Println("Migration completed")
}
