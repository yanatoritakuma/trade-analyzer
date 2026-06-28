package main

import (
	"context"
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

var ginLambda *ginadapter.GinLambdaV2

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

	// External（スタブ）
	claudeClient := external.NewClaudeClient()
	lineClient := external.NewLineClient()

	// Usecases
	userUsecase := usecase.NewUserUsecase(uow, userRepo)
	portfolioUsecase := usecase.NewPortfolioUsecase(tradeRepo, stockPriceRepo)
	positionUsecase := usecase.NewPositionUsecase(positionRepo, userRepo)
	watchlistUsecase := usecase.NewWatchlistUsecase(uow, watchlistRepo)
	tradeUsecase := usecase.NewTradeUsecase(tradeRepo)
	analysisUsecase := usecase.NewAnalysisUsecase(analysisRepo, settingRepo, claudeClient, lineClient)
	reportUsecase := usecase.NewReportUsecase(learningRepo, tradeRepo)
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

// Handler はAWS Lambda（API Gateway HTTP API）用のエントリポイント。
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env を読み込めませんでした（環境変数を直接使用します）")
	}

	r := setupRouter()

	if _, ok := os.LookupEnv("LAMBDA_TASK_ROOT"); ok {
		ginLambda = ginadapter.NewV2(r)
		lambda.Start(Handler)
	} else {
		log.Fatal(r.Run(":8080"))
	}
}
