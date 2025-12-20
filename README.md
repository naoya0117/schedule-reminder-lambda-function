# スケジュールリマインダー

Notionベースのスケジュールリマインダーサービス。Discord、LINE、Slackへの通知に対応。

## 概要

このAWS Lambda関数は以下の機能を提供します：

- 毎日指定時刻に自動実行（デフォルト: 09:00 JST）
- Notionの親データベースからリマインダー設定を読み込み
- 複数の子Notionデータベースからスケジュールを取得
- リマインド時期になったら通知を送信（1日前、N営業日前など）

## アーキテクチャ

- **言語**: Go 1.21
- **フレームワーク**: AWS SAM (Serverless Application Model)
- **データソース**: Notion API
- **通知チャネル**: Discord、LINE（Slackは近日対応予定）

## 必要な準備

1. **Notion Integration**
   - <https://www.notion.so/my-integrations> でIntegrationを作成
   - APIキー（`secret_`で始まる）を取得
   - データベースをIntegrationと共有

2. **AWSアカウント**
   - AWS CLIの設定
   - SAM CLIのインストール

3. **通知用設定**
   - Discord/Slack Webhook URL
   - LINE Messaging APIのチャネルアクセストークンと送信先ID

## 主な機能

- ✅ **柔軟なリマインドタイミング**: スケジュールごとに複数のリマインド時期を設定可能（1日前、4営業日前など）
- ✅ **営業日計算**: 営業日ベースのリマインドは自動的に週末・祝日をスキップ
- ✅ **複数の通知チャネル**: Discord、LINE対応（Slackは近日対応）
- ✅ **カスタマイズ可能なメッセージテンプレート**: 変数を使って通知メッセージをカスタマイズ
- ✅ **複数データベース対応**: 異なる設定で複数のNotionデータベースを監視
- ✅ **タイムゾーン対応**: Asia/Tokyo固定
- ✅ **コード変更不要**: すべての設定はNotion上で完結

## 動作の流れ

```
┌─────────────────────────────────────────────────────────┐
│       AWS Lambda（毎日定時実行）                         │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  1. 親Notionデータベースから設定を読み込み               │
│     （リマインダー設定マスター）                         │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  2. 各設定について：                                     │
│     - 子データベースからスケジュールを取得               │
│     - リマインド日を計算                                 │
│     - 今日送信すべきか判定                               │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  3. Discord/LINE/Slackへ通知を送信                      │
└─────────────────────────────────────────────────────────┘
```

## デプロイと使い方（まとめ）

### デプロイ手順

1. Notion Integrationを作成し、親/子データベースを共有する
2. 親DB（リマインダー設定マスター）と子DB（スケジュールDB）を作成する
   - CLIで作成する場合は `go run ./cmd/notion-init --api-key ... --parent-page-id ...`
3. 親DBのIDを `REMINDER_CONFIG_DB_ID` として控える
4. Notion APIキーを `NOTION_API_KEY` として用意する
5. Parameter Storeに登録して `sam deploy --guided` でデプロイする

### 使い方

1. 親DBにリマインダー設定を追加する
   - 必須: `名前`, `有効`, `対象データベースID`, `リマインドタイミング`, `通知チャネル`
2. 子DBにスケジュール/タスクを追加する
   - 必須: `タイトル`, `期限日`
3. 必要に応じて子DB側で上書きする
   - `リマインドタイミング`: レコード単位で通知タイミングを変更
   - `メッセージテンプレート`: レコード単位で通知文を変更
   - `説明`: `{description}` に差し込まれる補足情報
4. Lambdaは毎日定時に実行され、当日送るべき通知のみを送信する

## クイックスタート

### 1. Notionデータベースの設定

#### 親データベース: リマインダー設定マスター

以下のプロパティを持つデータベースを作成してください：

| プロパティ名 | タイプ | 必須 | 説明 |
|------------|--------|------|------|
| 名前 | Title | ✓ | リマインダー名 |
| 有効 | Checkbox | ✓ | このリマインダーを有効にするか |
| 対象データベースID | Text | ✓ | 監視対象の子データベースのID |
| リマインドタイミング | Multi-select | ✓ | リマインド時期（例: "1日前", "4営業日前"） |
| 通知チャネル | Select | ✓ | "Discord", "LINE", "Slack" のいずれか |
| Webhook URL | URL | * | Discord/Slack用のWebhook URL |
| チャネルアクセストークン | Text | * | LINE Messaging APIのチャネルアクセストークン |
| LINE送信先ID | Text | * | LINEの送信先ID（ユーザー/グループ/ルームID） |

**リマインドタイミング** の選択肢：

- `当日`
- `1日前`, `2日前`, `3日前`, `7日前`
- `1営業日前`, `2営業日前`, `3営業日前`, `4営業日前`, `5営業日前`
- `1週間前`, `2週間前`

#### 子データベース: スケジュール・タスク管理DB

各リマインダー設定は子データベースを参照します。子データベースには以下が必要です：

- **タイトル**プロパティ（固定）
- **期限日**プロパティ（固定）
- **説明**プロパティ（任意、通知テンプレートの `{description}` に入る）
- **リマインドタイミング**プロパティ（任意、各レコードでリマインド時期を上書き）
- **メッセージテンプレート**プロパティ（任意、各レコードでメッセージを上書き）

#### 「説明」プロパティの使い方

スケジュールDBの「説明」は、通知メッセージの `{description}` に差し込まれる補足情報です。
テンプレートで `{description}` を使わない限り通知文には出ません。

例:

```
説明: "定例ミーティングです。議題は前日までに追加してください。"
メッセージテンプレート: "【リマインド】{title}\n{description}\n{url}"
```

#### CLIで自動作成する場合

Notion APIを使って親/子データベースをまとめて作成できます。

```bash
cd src/app
go run ./cmd/notion-init \
  --api-key "secret_xxxxx" \
  --parent-page-id "your_parent_page_id"
```

サンプルの設定行も追加する場合は、Webhook URLを指定して実行します。

```bash
go run ./cmd/notion-init \
  --api-key "secret_xxxxx" \
  --parent-page-id "your_parent_page_id" \
  --create-sample-config \
  --sample-webhook-url "https://discord.com/api/webhooks/..."
```

### 2. データベースIDの取得

**親データベース（リマインダー設定マスター）の場合：**

1. ブラウザで親データベースを開く
2. URLからIDをコピー：

   ```
   https://www.notion.so/<DATABASE_ID>?v=...
                         ^^^^^^^^^^^^^^^^
                         これがREMINDER_CONFIG_DB_IDです
   ```

3. このIDを保存（AWSデプロイ時に必要）

**子データベース（スケジュール・タスクDB）の場合：**

1. ブラウザで各子データベースを開く
2. URLからIDをコピー（上記と同じ形式）
3. このIDを親データベースの"対象データベースID"プロパティに貼り付け

**例：**

```
親DB ID:   a1b2c3d4e5f67890abcdef1234567890
子DB 1 ID: f9e8d7c6b5a43210fedcba0987654321
子DB 2 ID: 1234567890abcdef1234567890abcdef
```

### 3. Parameter Storeに機密情報を設定

アプリケーションはAWS Systems Manager Parameter Storeから機密情報を読み取ります。

```bash
# Notion APIキーを登録
aws ssm put-parameter \
  --name "/lambda-functions/schedule-reminder/param-notion-api-key" \
  --value "secret_your_notion_api_key" \
  --type "SecureString"

# リマインダー設定マスターDBのIDを登録
aws ssm put-parameter \
  --name "/lambda-functions/schedule-reminder/param-reminder-config-db-id" \
  --value "your_parent_database_id" \
  --type "String"
```

### 4. AWSへデプロイ

```bash
cd src

# ビルド
sam build

# デプロイ
sam deploy --guided

# 以下の項目を入力：
# - Stack Name: schedule-reminder
# - AWS Region: ap-northeast-1（または任意のリージョン）
```

### 5. ローカルテスト（オプション）

#### LocalStack (docker-compose) を使用

```bash
# .envファイルを作成（.env.exampleをコピー）
cp src/.env.example src/.env

# .envファイルを編集してNotionの情報を設定
# PARAM_NOTION_API_KEY=secret_xxxxx
# PARAM_REMINDER_CONFIG_DB_ID=xxxxx

# Docker Composeで起動
docker-compose up -d

# LocalStackにParameter Storeの値が自動登録されます
# PARAM_プレフィックスの環境変数は以下のように変換されます：
# PARAM_NOTION_API_KEY → /lambda-functions/schedule-reminder/param-notion-api-key
```

#### SAM CLIでローカル実行（Parameter Store不要）

```bash
# 環境変数を設定（フォールバック用）
export NOTION_API_KEY="secret_your_notion_api_key"
export REMINDER_CONFIG_DB_ID="your_parent_database_id"

# ビルド
sam build

# ローカルで実行
sam local invoke ScheduleReminderFunction
```

## 設定例

### 例1: 毎日のミーティングリマインド

**親DB（設定）：**

```
名前: "朝会リマインド"
有効: ✓
対象データベースID: "abc123def456..."
リマインドタイミング: ["当日"]
通知チャネル: "Discord"
Webhook URL: "https://discord.com/api/webhooks/..."
```

**子DB（ミーティングスケジュール）：**

```
タイトル: "朝会"
期限日: 2025-12-01 10:00
Participants: [@チーム全員]
リマインドタイミング: ["当日"]
メッセージテンプレート: "【本日のミーティング】{title}\n時間: {due_date}\n{url}"
```

### 例2: タスク期限リマインド

**親DB（設定）：**

```
名前: "タスク期限リマインド"
有効: ✓
対象データベースID: "def456ghi789..."
リマインドタイミング: ["4営業日前", "1営業日前"]
通知チャネル: "Discord"
Webhook URL: "https://discord.com/api/webhooks/..."
```

**子DB（タスク管理）：**

```
タイトル: "API仕様書作成"
期限日: 2025-12-10
Status: "進行中"
Priority: "高"
Assignee: @山田
```

## メッセージテンプレート変数

テンプレートで使用可能な変数：

| 変数 | 説明 | 例 |
|------|------|-----|
| `{title}` | スケジュール・タスクのタイトル | "週次ミーティング" |
| `{due_date}` | 期限日 | "2025-12-01" |
| `{days_text}` | あと何日か | "明日" / "3日後" |
| `{url}` | NotionページのURL | "<https://notion.so/>..." |
| `{description}` | 説明（スケジュールDBの「説明」） | "四半期目標の確認" |

デフォルトテンプレート（指定なしの場合）：

```
【リマインド】{title}
期限: {due_date} ({days_text})
{url}
```

## プロジェクト構造

```
src/app/
├── main.go                                  # Lambda handler
├── go.mod                                   # Go依存関係
├── internal/
│   ├── domain/
│   │   ├── model/                          # ドメインモデル
│   │   │   ├── config.go                   # リマインダー設定
│   │   │   ├── schedule.go                 # スケジュール・タスク
│   │   │   └── notification.go             # 通知
│   │   ├── calculator/                     # 日付計算ロジック
│   │   │   ├── businessday.go              # 営業日計算
│   │   │   └── reminder.go                 # リマインド日計算
│   │   └── service/
│   │       ├── reminder.go                 # コアビジネスロジック
│   │       └── template.go                 # メッセージテンプレート
│   └── infrastructure/
│       ├── notion/                         # Notion APIクライアント
│       │   ├── client.go                   # 設定読み込み
│       │   └── schedule.go                 # スケジュール取得
│       └── notifier/                       # 通知送信
│           ├── notifier.go                 # インターフェース
│           ├── discord.go                  # Discord実装
│           └── factory.go                  # Notifierファクトリー
```

## 開発

### ビルド

```bash
cd src
sam build
```

### テスト

```bash
# ユニットテスト実行
cd src/app
go test ./...

# ローカルテスト
sam local invoke ScheduleReminderFunction
```

### 依存関係の更新

```bash
cd src/app
go get -u
go mod tidy
```

## ステップバイステップ セットアップガイド

### Step 1: Notion Integrationの作成

1. <https://www.notion.so/my-integrations> にアクセス
2. "New integration"をクリック
3. 名前を入力（例: "スケジュールリマインダー"）
4. ワークスペースを選択
5. "Submit"をクリック
6. "Internal Integration Token"をコピー（`secret_`で始まる）

### Step 2: Notionで親データベースを作成

1. Notionで新しいデータベースを作成
2. 名前を"リマインダー設定マスター"（任意の名前でOK）
3. 以下のプロパティを追加：

   | プロパティ | タイプ | オプション |
   |----------|--------|-----------|
   | 名前 | Title | - |
   | 有効 | Checkbox | - |
   | 対象データベースID | Text | - |
   | リマインドタイミング | Multi-select | `当日`, `1日前`, `2日前`, `1営業日前`等 |
   | 通知チャネル | Select | `Discord`, `LINE`, `Slack` |
   | Webhook URL | URL | - |
   | チャネルアクセストークン | Text | - |
   | LINE送信先ID | Text | - |

4. このデータベースをIntegrationと共有：
   - "共有"ボタンをクリック
   - Integration名で検索
   - "招待"をクリック

### Step 3: 子データベースの作成

1. スケジュール・タスク用のデータベースを作成（例: "ミーティングスケジュール"、"タスク管理"）
2. 各データベースに以下を含める：
   - タイトルプロパティ
   - 期限日用の日付プロパティ
   - リマインドタイミングプロパティ（任意）
   - メッセージテンプレートプロパティ（任意）
3. 各データベースをIntegrationと共有（Step 2.4と同じ手順）

### Step 4: 親データベースでリマインダーを設定

親データベースに新しいページを追加：

```
名前: "週次ミーティングリマインド"
有効: ✓
対象データベースID: [子データベースのIDを貼り付け]
リマインドタイミング: ["1日前", "当日"]
通知チャネル: "Discord"
Webhook URL: "https://discord.com/api/webhooks/..."
```

メッセージテンプレートとリマインド時期は、子データベースの各レコードで上書きできます。

### Step 5: Discord Webhook URLの取得

1. Discordを開く
2. サーバー設定 → 連携サービス → Webhookに移動
3. "ウェブフックを作成"をクリック
4. Webhook URLをコピー
5. Notion親データベースの"Webhook URL"欄に貼り付け

### Step 6: AWSへデプロイ

```bash
cd src

# 依存関係のインストール
cd app
go mod download
go mod tidy
cd ..

# ビルド
sam build

# デプロイ（初回）
sam deploy --guided

# 以下を入力：
# - Stack Name: schedule-reminder
# - AWS Region: ap-northeast-1（または希望のリージョン）
# - Parameter NotionAPIKey: secret_xxxxx（NotionのAPIキー）
# - Parameter ReminderConfigDBID: xxxxx（親データベースのID）
# - Confirm changes before deploy: Y
# - Allow SAM CLI IAM role creation: Y
# - Save arguments to samconfig.toml: Y

# 2回目以降のデプロイ（プロンプトなし）
sam deploy
```

### Step 7: デプロイの確認

1. AWS Lambdaコンソールで関数を確認
2. CloudWatch Eventsでスケジュールルールを確認
3. 手動で関数を実行してテスト：

   ```bash
   sam local invoke ScheduleReminderFunction
   ```

## 環境変数

Lambda関数は以下の環境変数を使用します：

| 変数 | 必須 | 説明 | 例 |
|------|------|------|-----|
| `NOTION_API_KEY` | ✓ | Notion Integration APIキー | `secret_xxxxx...` |
| `REMINDER_CONFIG_DB_ID` | ✓ | 親データベースのID | `a1b2c3d4e5f6...` |

これらはデプロイ時にCloudFormationパラメータから自動設定されます。

## トラブルシューティング

### "failed to query master database"エラー

**原因：**

- Notion APIキーが間違っているか期限切れ
- 親データベースがIntegrationと共有されていない
- データベースIDが間違っている

**解決方法：**

1. AWS LambdaのEnvironment variablesでNotion APIキーを確認
2. Notionで親データベースを開き、"共有"をクリック → Integrationが追加されているか確認
3. URLからデータベースIDを再確認

### 「対象データベースIDが必須」エラー

**原因：**

- 親データベースに"対象データベースID"プロパティがない
- 有効なリマインダー設定でプロパティが空

**解決方法：**

1. "対象データベースID"プロパティがTextタイプで存在することを確認
2. "Enabled"がチェックされている全ての行でIDが入力されているか確認
3. 子データベースのIDがURLから正しくコピーされているか確認

### 通知が送信されない

**原因：**

- Webhook URLが間違っている
- Discordサーバーの権限設定
- ネットワーク・タイムアウトの問題

**解決方法：**

1. Webhook URLを手動でテスト：

   ```bash
   curl -X POST "YOUR_WEBHOOK_URL" \
     -H "Content-Type: application/json" \
     -d '{"content": "テストメッセージ"}'
   ```

2. CloudWatch Logsで詳細なエラーメッセージを確認
3. リマインドタイミングが正しく設定されているか確認
4. 子データベースの期限日が未来の日付か確認

### "validation error: DueDate: required"エラー

**原因：**

- 子データベースのレコードに日付プロパティがない
- 子DBに"タイトル"と"期限日"プロパティがない

**解決方法：**

1. 子データベースの全レコードで日付プロパティが入力されているか確認
2. 子DBに"タイトル"と"期限日"プロパティが存在するか確認
3. プロパティ名は大文字小文字を区別することに注意

### Lambdaタイムアウトエラー

**原因：**

- 処理するスケジュールが多すぎる
- Notion APIのレスポンスが遅い
- 送信するリマインドが多すぎる

**解決方法：**

1. `template.yaml`でタイムアウトを増やす（現在60秒）
2. リマインダーを複数の設定に分割
3. フィルターでスケジュール数を減らして最適化

## 監視

### CloudWatch Logs

デバッグ用のログを表示：

```bash
aws logs tail /aws/lambda/schedule-reminder-ScheduleReminderFunction-xxx --follow
```

### 主要なログメッセージ

- `=== Schedule Reminder Lambda Started ===` - 関数開始
- `Loaded X reminder configurations` - 設定読み込み成功
- `Processing: [Name]` - 特定のリマインダーを処理中
- `Found X schedules` - 見つかったスケジュール数
- `✓ Sent Discord notification` - 通知送信成功
- `=== Schedule Reminder Lambda Completed ===` - 関数完了

### エラーパターン

ログで以下のパターンを探す：

- `failed to query master database` → Notionアクセスの問題
- `failed to fetch schedules` → 子データベースアクセスの問題
- `failed to send notification` → Webhook・ネットワークの問題
- `validation error` → 必須フィールドの欠落

## 高度な使い方

### カスタムメッセージテンプレート

子データベースの各レコードの"メッセージテンプレート"に設定し、任意のプロパティをテンプレートで使用：

```
メッセージテンプレート: "📅 {title}\n期限: {due_date}\n優先度: {priority}\n担当: {assignee}\n\n{description}\n\n詳細: {url}"
```

### 複数の通知チャネル

同じ子データベースに対して異なるチャネルで複数のリマインダーを設定可能：

**リマインダー1:**

- 名前: "タスクリマインド（チーム）"
- Channel: Discord
- Webhook: チームチャンネルのwebhook

**リマインダー2:**

- 名前: "タスクリマインド（個人）"
- Channel: LINE
- Token: 個人用LINEトークン
- LINE送信先ID: 個人のユーザーID

### 営業日の例

**4営業日前のリマインド：**

- 期限日: 12月15日（金）
- リマインド日: 12月11日（月）（週末をスキップ）

**1営業日前のリマインド：**

- 期限日: 12月11日（月）
- リマインド日: 12月8日（金）

## ロードマップ

### Phase 1（現在 - MVP）

- [x] Discord通知対応
- [x] 基本的なリマインドタイミング（〜日前）
- [x] 営業日計算
- [x] Notionデータベース連携
- [x] カスタマイズ可能なメッセージテンプレート

### Phase 2（次期）

- [x] LINE通知対応
- [ ] Slack通知対応
- [ ] 祝日API連携（自動祝日読み込み）
- [ ] 通知履歴管理（重複防止）
- [ ] 失敗時のリトライロジック

### Phase 3（将来）

- [ ] 設定管理用Web UI
- [ ] スケジュールごとの複数タイムゾーン対応
- [ ] リッチフォーマット（Discord embeds、LINE Flex Messages）
- [ ] 通知分析ダッシュボード
- [ ] SMS・Email通知対応
