import * as path from 'path';
import * as fs from 'fs';
import { Construct } from 'constructs';
import {
  Stack,
  StackProps,
  Duration,
  RemovalPolicy,
  CfnOutput,
  aws_lambda as lambda,
  aws_apigatewayv2 as apigwv2,
  aws_apigatewayv2_integrations as integrations,
  aws_s3 as s3,
  aws_events as events,
  aws_events_targets as targets,
  aws_secretsmanager as secretsmanager,
  aws_logs as logs,
} from 'aws-cdk-lib';

/**
 * TradeAnalyzerStack は infra_requirements.md（1章 全体アーキテクチャ）の AWS 側を構築する。
 *
 *   ブラウザ → API Gateway(HTTP API) → Go Lambda(back/) → Neon(PostgreSQL・AWS外)
 *   EventBridge(定期実行) → 分析/週次の内部エンドポイント(API Destination)
 *   S3(学習CSVのバージョン管理)
 *
 * Neon は AWS リソースではないため本スタックでは管理しない（接続URLはシークレット経由でLambdaに渡す）。
 * Python 株価取得Lambda（lambda/）は別フェーズのため、既存LambdaのARNを context で渡したときのみ配線する。
 */
export class TradeAnalyzerStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    // ---- context（cdk.json / -c で上書き可能） ----
    const appSecretName = this.node.tryGetContext('appSecretName') as string;
    const frontendOrigin = (this.node.tryGetContext('frontendOrigin') as string) ?? '';
    const anthropicModel = (this.node.tryGetContext('anthropicModel') as string) ?? '';
    const anthropicModelWeekly = (this.node.tryGetContext('anthropicModelWeekly') as string) ?? '';
    const jwtAccessExpireHours = (this.node.tryGetContext('jwtAccessExpireHours') as string) ?? '24';
    const jwtRefreshExpireDays = (this.node.tryGetContext('jwtRefreshExpireDays') as string) ?? '30';
    const stockFetchLambdaArn = (this.node.tryGetContext('stockFetchLambdaArn') as string) ?? '';

    // ---- Go Lambda のビルド成果物（make build-lambda で生成）----
    const lambdaZip = path.join(__dirname, '..', '..', 'back', 'lambda-handler.zip');
    if (!fs.existsSync(lambdaZip)) {
      throw new Error(
        `Go Lambda の成果物が見つかりません: ${lambdaZip}\n` +
          'デプロイ前に back/ で `make build-lambda` を実行してください。',
      );
    }

    // ---- アプリのシークレット（Secrets Manager・事前に手動作成）----
    // 1つのシークレットにJSONで全キーを格納する想定:
    //   NEON_DATABASE_URL, JWT_SECRET, INTERNAL_API_SECRET,
    //   ANTHROPIC_API_KEY, LINE_CHANNEL_ACCESS_TOKEN, LINE_USER_ID
    const appSecret = secretsmanager.Secret.fromSecretNameV2(this, 'AppSecret', appSecretName);

    // ---- S3: 学習CSVのバージョン管理（learning_versions のソース）----
    const learningBucket = new s3.Bucket(this, 'LearningBucket', {
      versioned: true,
      encryption: s3.BucketEncryption.S3_MANAGED,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      enforceSSL: true,
      removalPolicy: RemovalPolicy.RETAIN, // 学習データは誤削除を防ぐため保持
    });

    // ---- Lambda 共通環境変数 ----
    // 注意: AWS_REGION は Lambda が自動付与する予約環境変数のため設定しない。
    //       DATABASE_URL は未設定にして NEON_DATABASE_URL を使わせる（db.go の解決順）。
    //       シークレットは CloudFormation の dynamic reference でデプロイ時に解決される。
    const commonEnv: { [key: string]: string } = {
      APP_ENV: 'production',
      FRONTEND_ORIGIN: frontendOrigin,
      S3_BUCKET_NAME: learningBucket.bucketName,
      JWT_ACCESS_EXPIRE_HOURS: jwtAccessExpireHours,
      JWT_REFRESH_EXPIRE_DAYS: jwtRefreshExpireDays,
      ...(anthropicModel ? { ANTHROPIC_MODEL: anthropicModel } : {}),
      ...(anthropicModelWeekly ? { ANTHROPIC_MODEL_WEEKLY: anthropicModelWeekly } : {}),
      NEON_DATABASE_URL: appSecret.secretValueFromJson('NEON_DATABASE_URL').unsafeUnwrap(),
      JWT_SECRET: appSecret.secretValueFromJson('JWT_SECRET').unsafeUnwrap(),
      INTERNAL_API_SECRET: appSecret.secretValueFromJson('INTERNAL_API_SECRET').unsafeUnwrap(),
      ANTHROPIC_API_KEY: appSecret.secretValueFromJson('ANTHROPIC_API_KEY').unsafeUnwrap(),
      LINE_CHANNEL_ACCESS_TOKEN: appSecret.secretValueFromJson('LINE_CHANNEL_ACCESS_TOKEN').unsafeUnwrap(),
      LINE_USER_ID: appSecret.secretValueFromJson('LINE_USER_ID').unsafeUnwrap(),
    };

    // 同一のGoバイナリ(bootstrap)を2つのLambdaにデプロイする。
    // main.go の dispatch がイベント種別で「API Gatewayリクエスト」と「定期バッチ(job)」を振り分ける。
    const codeAsset = lambda.Code.fromAsset(lambdaZip);

    // ---- ① API用Lambda（API Gateway HTTP API のバックエンド）----
    // ユーザー向け同期APIのため短いタイムアウト（30秒）で隔離する。
    const apiLogGroup = new logs.LogGroup(this, 'ApiFunctionLogs', {
      retention: logs.RetentionDays.TWO_WEEKS,
      removalPolicy: RemovalPolicy.DESTROY,
    });
    const goFn = new lambda.Function(this, 'ApiFunction', {
      runtime: lambda.Runtime.PROVIDED_AL2023, // Goカスタムランタイム
      handler: 'bootstrap',
      code: codeAsset,
      architecture: lambda.Architecture.X86_64, // make build-lambda は GOARCH=amd64
      memorySize: 256,
      timeout: Duration.seconds(30), // API Gateway HTTP API の上限に合わせる
      logGroup: apiLogGroup,
      environment: commonEnv,
    });

    // ---- ② Worker用Lambda（EventBridge定期バッチ：分析・週次レポート）----
    // 多銘柄のClaude分析が長時間化し得るため、API Gatewayを介さず直接呼び出し、最大15分まで許容する。
    const workerLogGroup = new logs.LogGroup(this, 'WorkerFunctionLogs', {
      retention: logs.RetentionDays.TWO_WEEKS,
      removalPolicy: RemovalPolicy.DESTROY,
    });
    const workerFn = new lambda.Function(this, 'WorkerFunction', {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      code: codeAsset,
      architecture: lambda.Architecture.X86_64,
      memorySize: 512,
      timeout: Duration.minutes(15), // 定期バッチは最大15分（API Gatewayの30秒制限を受けない）
      logGroup: workerLogGroup,
      environment: commonEnv,
    });

    // 学習CSVのPUT/参照を許可（週次レポートがS3アップロードを行う）
    learningBucket.grantReadWrite(goFn);
    learningBucket.grantReadWrite(workerFn);

    // ---- API Gateway（HTTP API・payload v2）----
    // gin の全ルート(/api, /internal, /health)を1つのLambdaにプロキシする（OPTIONSも含む）。
    // CORS は API Gateway 側では設定しない。理由: Goアプリ(gin-contrib/cors)が
    // FRONTEND_ORIGIN を許可オリジンとして CORS ヘッダを付与しており、API Gateway 側でも
    // 付与すると Access-Control-Allow-Origin が重複してブラウザに拒否されるため、アプリ側に一本化する。
    const httpApi = new apigwv2.HttpApi(this, 'HttpApi', {
      description: 'trade-analyzer API (proxy to Go Lambda)',
      defaultIntegration: new integrations.HttpLambdaIntegration('GoIntegration', goFn, {
        payloadFormatVersion: apigwv2.PayloadFormatVersion.VERSION_2_0,
      }),
    });

    const apiBaseUrl = httpApi.apiEndpoint; // https://{id}.execute-api.{region}.amazonaws.com

    // ---- EventBridge: 定期実行（時刻はJST→UTC換算）----
    // 分析・週次レポートは Worker Lambda を「直接呼び出し」する（API Gateway/API Destinationを介さない）。
    // 利点: ①EventBridge Connectionが不要 → 付随するSecrets Managerシークレット($0.40/月)が発生しない
    //       ②API Gatewayの30秒制限を受けず最大15分まで実行できる
    // 認証はIAM（EventBridgeがWorker Lambdaをinvokeする権限）で担保され、X-Internal-Secretは不要。
    // 定数input {"job": "..."} を main.go の dispatch が判定して対応usecaseを実行する。

    // 分析実行: 平日 15:30 JST = 06:30 UTC（Mon-Fri）
    new events.Rule(this, 'AnalyzeSchedule', {
      description: '分析実行（平日15:30 JST）',
      schedule: events.Schedule.cron({ minute: '30', hour: '6', weekDay: 'MON-FRI' }),
      targets: [
        new targets.LambdaFunction(workerFn, {
          event: events.RuleTargetInput.fromObject({ job: 'analyze' }),
        }),
      ],
    });

    // 週次レポート: 日曜 18:00 JST = 09:00 UTC（Sun）
    new events.Rule(this, 'WeeklyReportSchedule', {
      description: '週次レポート生成（日曜18:00 JST）',
      schedule: events.Schedule.cron({ minute: '0', hour: '9', weekDay: 'SUN' }),
      targets: [
        new targets.LambdaFunction(workerFn, {
          event: events.RuleTargetInput.fromObject({ job: 'weekly_report' }),
        }),
      ],
    });

    // 株価取得: 平日 15:00 JST = 06:00 UTC（Mon-Fri）→ Python株価取得Lambda
    // 別フェーズで lambda/ を実装・デプロイ後、そのARNを context(stockFetchLambdaArn)で渡すと配線される。
    if (stockFetchLambdaArn) {
      const stockFetchFn = lambda.Function.fromFunctionArn(this, 'StockFetchFunction', stockFetchLambdaArn);
      new events.Rule(this, 'StockFetchSchedule', {
        description: '株価取得（平日15:00 JST）',
        schedule: events.Schedule.cron({ minute: '0', hour: '6', weekDay: 'MON-FRI' }),
        targets: [new targets.LambdaFunction(stockFetchFn)],
      });
    }

    // ---- 出力 ----
    new CfnOutput(this, 'ApiBaseUrl', {
      value: apiBaseUrl,
      description: 'フロントの NEXT_PUBLIC_API_BASE_URL に設定するAPIのベースURL',
    });
    new CfnOutput(this, 'LearningBucketName', {
      value: learningBucket.bucketName,
      description: '学習CSV保存先のS3バケット名（Lambdaの S3_BUCKET_NAME に自動設定済み）',
    });
  }
}
