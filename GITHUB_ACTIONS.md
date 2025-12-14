# GitHub Actions セットアップガイド

このプロジェクトは GitHub Actions を使用した自動デプロイに対応しています。

## 前提条件

1. AWSアカウント
2. GitHub リポジトリ
3. AWS OIDC プロバイダーの設定（推奨）

## セットアップ手順

### 1. AWS OIDC プロバイダーの作成

GitHub Actions から AWS にアクセスするために、OIDC プロバイダーを設定します。

```bash
# OIDC プロバイダーを作成
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

### 2. IAM ロールの作成

GitHub Actions が使用する IAM ロールを作成します。

**trust-policy.json** を作成：

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_GITHUB_USERNAME/aws-lambda-functions:*"
        }
      }
    }
  ]
}
```

**ロールを作成：**

```bash
# IAM ロールを作成
aws iam create-role \
  --role-name ScheduleReminderGitHubDeployRole \
  --assume-role-policy-document file://trust-policy.json

# 必要な権限ポリシーをアタッチ
aws iam attach-role-policy \
  --role-name ScheduleReminderGitHubDeployRole \
  --policy-arn arn:aws:iam::aws:policy/AWSCloudFormationFullAccess

aws iam attach-role-policy \
  --role-name ScheduleReminderGitHubDeployRole \
  --policy-arn arn:aws:iam::aws:policy/IAMFullAccess

aws iam attach-role-policy \
  --role-name ScheduleReminderGitHubDeployRole \
  --policy-arn arn:aws:iam::aws:policy/AWSLambda_FullAccess

aws iam attach-role-policy \
  --role-name ScheduleReminderGitHubDeployRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess

aws iam attach-role-policy \
  --role-name ScheduleReminderGitHubDeployRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMFullAccess

aws iam attach-role-policy \
  --role-name ScheduleReminderGitHubDeployRole \
  --policy-arn arn:aws:iam::aws:policy/CloudWatchEventsFullAccess
```

### 3. GitHub Secrets の設定

GitHubリポジトリの Settings → Secrets and variables → Actions で以下のシークレットを設定します：

| Secret名 | 説明 | 例 |
|----------|------|-----|
| `AWS_ACCOUNT_ID` | AWSアカウントID | `123456789012` |
| `PARAM_NOTION_API_KEY` | Notion Integration APIキー | `secret_xxxxx...` |
| `PARAM_REMINDER_CONFIG_DB_ID` | リマインダー設定マスターDBのID | `a1b2c3d4e5f6...` |

### 4. SAM設定ファイルの作成

`src/samconfig.toml` を作成：

```toml
version = 0.1

[prod]
[prod.deploy]
[prod.deploy.parameters]
stack_name = "schedule-reminder"
s3_prefix = "schedule-reminder"
region = "ap-northeast-1"
confirm_changeset = false
capabilities = "CAPABILITY_IAM"
disable_rollback = false
image_repositories = []
resolve_s3 = true

[prod.build]
[prod.build.parameters]
```

### 5. ワークフローの動作確認

1. `main` または `master` ブランチにプッシュするか、GitHub UI から手動でワークフローを実行
2. Actions タブで実行状況を確認
3. デプロイが完了すると、Parameter Store に値が自動的に登録されます

## ワークフローの動作

`.github/workflows/deploy.yml` は以下の処理を自動実行します：

1. **コードのチェックアウト**
2. **Go環境のセットアップ** (Go 1.21)
3. **SAM CLIのセットアップ**
4. **AWS認証情報の設定** (OIDC経由)
5. **SAMアプリケーションのビルド**
6. **SAMアプリケーションのデプロイ**
7. **Parameter Storeへの値の登録**
   - `/lambda-functions/schedule-reminder/param-notion-api-key`
   - `/lambda-functions/schedule-reminder/param-reminder-config-db-id`
8. **デプロイ情報の出力**

## トラブルシューティング

### "Error: User is not authorized to perform: sts:AssumeRoleWithWebIdentity"

**原因：** IAMロールの信頼ポリシーが正しく設定されていない

**解決方法：**
1. trust-policy.json の `YOUR_AWS_ACCOUNT_ID` と `YOUR_GITHUB_USERNAME` を確認
2. OIDCプロバイダーのARNが正しいか確認
3. リポジトリ名が正しいか確認

### "Error: Parameter Store access denied"

**原因：** IAMロールにSSMへのアクセス権限がない

**解決方法：**
```bash
aws iam attach-role-policy \
  --role-name ScheduleReminderGitHubDeployRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMFullAccess
```

### "samconfig.toml not found"

**原因：** SAM設定ファイルが存在しない

**解決方法：**
`src/samconfig.toml` を上記のテンプレートを使って作成してください。

## ローカル開発との併用

GitHub Actions を使用している場合でも、ローカル開発は可能です：

```bash
# .envファイルを作成
cp src/.env.example src/.env

# 環境変数を設定
# PARAM_NOTION_API_KEY=secret_xxxxx
# PARAM_REMINDER_CONFIG_DB_ID=xxxxx

# Docker Composeで起動
docker-compose up -d
```

ローカル環境では LocalStack が Parameter Store をシミュレートします。

## セキュリティのベストプラクティス

1. **最小権限の原則**: IAMロールには必要最小限の権限のみを付与
2. **シークレットの管理**: 
   - GitHub Secrets に機密情報を保存
   - Parameter Store で SecureString タイプを使用
3. **ブランチ保護**: 
   - main/master ブランチに直接プッシュを制限
   - プルリクエスト経由でのマージを必須化
4. **環境の分離**: 
   - 本番環境と開発環境で異なる AWS アカウントを使用（推奨）

## 参考資料

- [GitHub Actions - AWS認証](https://github.com/aws-actions/configure-aws-credentials)
- [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
- [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
