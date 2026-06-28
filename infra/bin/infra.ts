#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { TradeAnalyzerStack } from '../lib/trade-analyzer-stack';

const app = new cdk.App();

// account/region は環境変数（CDK_DEFAULT_*）から解決する。
// 例: AWS_PROFILE と CDK_DEFAULT_REGION=ap-northeast-1 を設定して deploy する。
const env: cdk.Environment = {
  account: process.env.CDK_DEFAULT_ACCOUNT,
  region: process.env.CDK_DEFAULT_REGION ?? 'ap-northeast-1',
};

new TradeAnalyzerStack(app, 'TradeAnalyzerStack', {
  env,
  description: 'trade-analyzer: Go Lambda + API Gateway HTTP API + EventBridge + S3 (ISSUE #7)',
});
