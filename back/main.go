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

	"github.com/yanatoritakuma/trade-analyzer/back/db"
	"github.com/yanatoritakuma/trade-analyzer/back/router"
)

var ginLambda *ginadapter.GinLambdaV2

func setupRouter() *gin.Engine {
	database := db.NewDB()
	return router.NewRouter(database)
}

// Handler はAWS Lambda（API Gateway HTTP API）用のエントリポイント。
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	// ローカルでは .env を読み込む（Lambda上では環境変数が直接渡るため無視してよい）
	if err := godotenv.Load(); err != nil {
		log.Println(".env を読み込めませんでした（環境変数を直接使用します）")
	}

	r := setupRouter()

	if _, ok := os.LookupEnv("LAMBDA_TASK_ROOT"); ok {
		// Lambda実行環境
		ginLambda = ginadapter.NewV2(r)
		lambda.Start(Handler)
	} else {
		// ローカル実行
		log.Fatal(r.Run(":8080"))
	}
}
