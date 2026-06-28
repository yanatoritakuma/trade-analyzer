package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/yanatoritakuma/trade-analyzer/back/controller"
	"github.com/yanatoritakuma/trade-analyzer/back/db"
	"github.com/yanatoritakuma/trade-analyzer/back/external"
	"github.com/yanatoritakuma/trade-analyzer/back/repository"
	"github.com/yanatoritakuma/trade-analyzer/back/router"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
)

var (
	ginLambda *ginadapter.GinLambdaV2

	// 定期実行（EventBridge → Lambda 直接呼び出し）から参照するユースケース。
	// API Gateway経路を通らないため X-Internal-Secret も不要（認証はIAM）。
	analysisUsecaseRef *usecase.AnalysisUsecase
	reportUsecaseRef   *usecase.ReportUsecase
)

func setupRouter() *gin.Engine {
	database := db.NewDB()

	// UoW（書き込み系トランザクション）
	uow := repository.NewUnitOfWork(database)

	// 参照系Repository
	userRepo := repository.NewUserRepositoryImpl(database)
	invitationRepo := repository.NewInvitationCodeRepositoryImpl(database)
	watchlistRepo := repository.NewWatchlistRepositoryImpl(database)
	candidateRepo := repository.NewCandidateRepositoryImpl(database)
	tradeRepo := repository.NewTradeRepositoryImpl(database)
	positionRepo := repository.NewPositionRepositoryImpl(database)
	stockPriceRepo := repository.NewStockPriceRepositoryImpl(database)
	analysisRepo := repository.NewAnalysisLogRepositoryImpl(database)
	themeRepo := repository.NewThemeRepositoryImpl(database)
	settingRepo := repository.NewSettingRepositoryImpl(database)
	learningRepo := repository.NewLearningLogRepositoryImpl(database)

	// External（環境変数が未設定ならスタブにフォールバック）
	// モデルはジョブ別に割り当てる:
	//   - 毎日分析(RunScheduled): ANTHROPIC_MODEL→既定Sonnet（高頻度・コスト重視）
	//   - 週次学習(RunWeekly):    ANTHROPIC_MODEL_WEEKLY→既定Opus 4.8（低頻度・高レバレッジ）
	dailyClaudeClient := external.NewClaudeClient()
	weeklyModel := os.Getenv("ANTHROPIC_MODEL_WEEKLY")
	if weeklyModel == "" {
		weeklyModel = "claude-opus-4-8"
	}
	weeklyClaudeClient := external.NewClaudeClientForModel(weeklyModel)
	lineClient := external.NewLineClient()

	// Usecases
	userUsecase := usecase.NewUserUsecase(uow, userRepo)
	portfolioUsecase := usecase.NewPortfolioUsecase(tradeRepo, stockPriceRepo)
	positionUsecase := usecase.NewPositionUsecase(positionRepo, userRepo)
	watchlistUsecase := usecase.NewWatchlistUsecase(uow, watchlistRepo)
	tradeUsecase := usecase.NewTradeUsecase(tradeRepo)
	analysisUsecase := usecase.NewAnalysisUsecase(analysisRepo, settingRepo, dailyClaudeClient, lineClient)
	reportUsecase := usecase.NewReportUsecase(learningRepo, tradeRepo, weeklyClaudeClient)

	// 定期実行（EventBridge直接呼び出し）のハンドラから参照できるよう保持する。
	analysisUsecaseRef = analysisUsecase
	reportUsecaseRef = reportUsecase
	adminUsecase := usecase.NewAdminUsecase(userRepo)
	invitationUsecase := usecase.NewInvitationUsecase(invitationRepo)
	themeUsecase := usecase.NewThemeUsecase(themeRepo)
	candidateUsecase := usecase.NewCandidateUsecase(uow, candidateRepo)
	internalUsecase := usecase.NewInternalUsecase(stockPriceRepo)

	// Controllers
	controllers := &router.Controllers{
		User:       controller.NewUserController(userUsecase),
		Portfolio:  controller.NewPortfolioController(portfolioUsecase, positionUsecase),
		Position:   controller.NewPositionController(positionUsecase),
		Watchlist:  controller.NewWatchlistController(watchlistUsecase),
		Trade:      controller.NewTradeController(tradeUsecase),
		Analysis:   controller.NewAnalysisController(analysisUsecase),
		Report:     controller.NewReportController(reportUsecase),
		Admin:      controller.NewAdminController(adminUsecase),
		Invitation: controller.NewInvitationController(invitationUsecase),
		Theme:      controller.NewThemeController(themeUsecase),
		Candidate:  controller.NewCandidateController(candidateUsecase),
		Internal:   controller.NewInternalController(watchlistUsecase, internalUsecase),
	}

	return router.NewRouter(database, controllers)
}

// scheduledEvent はEventBridgeルールが渡す定数input（{"job": "..."}）。
type scheduledEvent struct {
	Job string `json:"job"`
}

// dispatch はLambdaのエントリポイント。受け取ったイベントの形状で処理を振り分ける。
//   - EventBridge定期実行（{"job":"analyze"|"weekly_report"}）→ 対応するusecaseを直接実行
//   - それ以外（API Gateway HTTP API イベント）→ ginにプロキシ
//
// API GatewayのイベントにはトップレベルにJobフィールドが無いため、Jobが空なら
// API Gatewayリクエストとして扱う。これにより同一バイナリでAPIと定期バッチの両方を処理できる。
func dispatch(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var ev scheduledEvent
	if err := json.Unmarshal(raw, &ev); err == nil && ev.Job != "" {
		switch ev.Job {
		case "analyze":
			return nil, analysisUsecaseRef.RunScheduled(ctx)
		case "weekly_report":
			return nil, reportUsecaseRef.RunWeekly(ctx)
		default:
			return nil, fmt.Errorf("dispatch: 未知のjob %q", ev.Job)
		}
	}

	var req events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("dispatch: APIGatewayV2HTTPRequestのunmarshalに失敗: %w", err)
	}
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env を読み込めませんでした（環境変数を直接使用します）")
	}

	r := setupRouter()

	if _, ok := os.LookupEnv("LAMBDA_TASK_ROOT"); ok {
		ginLambda = ginadapter.NewV2(r)
		lambda.Start(dispatch)
	} else {
		log.Fatal(r.Run(":8080"))
	}
}
