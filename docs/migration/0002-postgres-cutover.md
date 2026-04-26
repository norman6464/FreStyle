# Phase 0b: PostgreSQL 切替手順 (cutover)

## 目的

既存 MariaDB (`frestyle-prod`) を **PostgreSQL 16** (`frestyle-prod-postgres`) に切替える。マルチテナント B2B SaaS の本格運用に備え、Row-Level Security と JSONB が使える土台に乗せる。

## 前提

- Phase 0 Step A (`007_multitenancy_foundation.sql`) は MariaDB 上で実行済み
- 既存データは FreStyle 社 (id=1) にバックフィル済み
- このセッションで作成した:
  - `FreStyle/src/main/resources/schema-postgres.sql` (PostgreSQL 用全 DDL)
  - `FreStyle/src/main/resources/application-postgres.properties` (Spring Profile)
  - `FreStyle/build.gradle` に `org.postgresql:postgresql` ドライバ追加
  - `frestyle-infrastructure/infrastructure/cloudformation/templates/runtime/rds-postgres.yml`
  - `frestyle-infrastructure/scripts/migrate-mariadb-to-postgres.sh`

## 切替の段階構成

```
┌──────────────────────────────────────────────────────────────────────────┐
│ Step B1: 新 PostgreSQL インスタンス作成（並行運用、既存 MariaDB は稼働継続）│
├──────────────────────────────────────────────────────────────────────────┤
│ Step B2: schema-postgres.sql をスキーマ定義として流す                      │
├──────────────────────────────────────────────────────────────────────────┤
│ Step B3: ECS タスク内でデータ移行スクリプト実行 (MariaDB → PostgreSQL)     │
├──────────────────────────────────────────────────────────────────────────┤
│ Step B4: ステージング ECS の DB_URL を Postgres に切替えて動作確認         │
├──────────────────────────────────────────────────────────────────────────┤
│ Step B5: 本番メンテナンスウィンドウ - MariaDB を read-only → 最終差分取込  │
│         → 本番 ECS の DB_URL を Postgres に切替 → 再デプロイ              │
├──────────────────────────────────────────────────────────────────────────┤
│ Step B6: 1 週間モニタリング → 問題なければ MariaDB スタックを destroy      │
└──────────────────────────────────────────────────────────────────────────┘
```

## Step B1: 新 PostgreSQL インスタンスを作成

```bash
cd ~/Desktop/FreStyle/frestyle-infrastructure

# 1) パスワードを Secrets Manager に保管 (既存とは別 Secret 名)
read -s -p "PG_MASTER_PASSWORD: " PG_PASS; echo
aws secretsmanager create-secret --region ap-northeast-1 \
  --name frestyle-prod/pg-master-password \
  --secret-string "$PG_PASS"

# 2) 既存 RDS SG に PostgreSQL ポート (5432) を追加
RDS_SG=$(aws cloudformation describe-stacks --region ap-northeast-1 \
  --stack-name frestyle-prod-sg \
  --query 'Stacks[0].Outputs[?OutputKey==`RdsSecurityGroupId`].OutputValue' --output text)
ECS_SG=$(aws cloudformation describe-stacks --region ap-northeast-1 \
  --stack-name frestyle-prod-sg \
  --query 'Stacks[0].Outputs[?OutputKey==`EcsServiceSecurityGroupId`].OutputValue' --output text)
aws ec2 authorize-security-group-ingress --region ap-northeast-1 \
  --group-id $RDS_SG \
  --ip-permissions "IpProtocol=tcp,FromPort=5432,ToPort=5432,UserIdGroupPairs=[{GroupId=$ECS_SG,Description='ECS to PostgreSQL'}]"

# 3) PostgreSQL CFn スタックをデプロイ
aws cloudformation deploy --region ap-northeast-1 \
  --stack-name frestyle-prod-rds-postgres \
  --template-file infrastructure/cloudformation/templates/runtime/rds-postgres.yml \
  --parameter-overrides \
    Environment=prod \
    RdsSecurityGroupId=$RDS_SG \
    DBMasterPassword="$PG_PASS" \
  --capabilities CAPABILITY_IAM \
  --no-fail-on-empty-changeset \
  --tags Project=FreStyle Environment=prod ManagedBy=CloudFormation

# 4) 確認
PG_HOST=$(aws cloudformation describe-stacks --region ap-northeast-1 \
  --stack-name frestyle-prod-rds-postgres \
  --query 'Stacks[0].Outputs[?OutputKey==`DBEndpoint`].OutputValue' --output text)
echo "PostgreSQL endpoint: $PG_HOST"
```

## Step B2: スキーマを PostgreSQL に流す

ECS execute-command で稼働中の Spring Boot コンテナから流す:

```bash
TASK_ID=$(aws ecs list-tasks --region ap-northeast-1 --cluster frestyle-prod \
  --service-name frestyle-prod-svc --desired-status RUNNING \
  --query 'taskArns[0]' --output text | awk -F/ '{print $NF}')

# postgresql-client インストール (一時)
aws ecs execute-command --region ap-northeast-1 \
  --cluster frestyle-prod --task $TASK_ID --container fre-style \
  --command "bash -c 'apt-get update -qq && apt-get install -y -qq postgresql-client && which psql'" \
  --interactive

# schema-postgres.sql を base64 で送って流す
SQL_B64=$(base64 -i FreStyle/src/main/resources/schema-postgres.sql | tr -d '\n')
PG_PASS=$(aws secretsmanager get-secret-value --region ap-northeast-1 \
  --secret-id frestyle-prod/pg-master-password --query 'SecretString' --output text)

aws ecs execute-command --region ap-northeast-1 \
  --cluster frestyle-prod --task $TASK_ID --container fre-style \
  --command "bash -c 'echo $SQL_B64 | base64 -d > /tmp/s.sql && PGPASSWORD=\"$PG_PASS\" psql -h $PG_HOST -U postgres -d fre_style -f /tmp/s.sql 2>&1 | tail -30'" \
  --interactive
```

## Step B3: データ移行スクリプト実行

```bash
# MariaDB / PostgreSQL の認証情報を環境変数で渡す
MA_HOST=frestyle-prod.ctwsumy6osqz.ap-northeast-1.rds.amazonaws.com
MA_PASS=$(aws secretsmanager get-secret-value --region ap-northeast-1 \
  --secret-id frestyle-prod/db-master-password --query 'SecretString' --output text)

# スクリプトを base64 で送って実行
SCR_B64=$(base64 -i frestyle-infrastructure/scripts/migrate-mariadb-to-postgres.sh | tr -d '\n')

aws ecs execute-command --region ap-northeast-1 \
  --cluster frestyle-prod --task $TASK_ID --container fre-style \
  --command "bash -c 'echo $SCR_B64 | base64 -d > /tmp/m.sh && chmod +x /tmp/m.sh && \
    MA_HOST=$MA_HOST MA_USER=admin MA_PASS=\"$MA_PASS\" \
    PG_HOST=$PG_HOST PG_USER=postgres PG_PASS=\"$PG_PASS\" \
    DB_NAME=fre_style /tmp/m.sh 2>&1 | tail -100'" \
  --interactive
```

## Step B4: ECS の DB_URL を切替えて検証

```bash
# 新しい接続文字列を Secrets Manager に保管
aws secretsmanager create-secret --region ap-northeast-1 \
  --name frestyle-prod/db-url-postgres \
  --secret-string "jdbc:postgresql://$PG_HOST:5432/fre_style"

# ECS Task Definition の env を更新（新リビジョンを作る）
# - SPRING_PROFILES_ACTIVE=postgres
# - DB_URL=jdbc:postgresql://...
# - DB_USER=postgres
# - DB_PASS=(Secrets Manager 参照)
# 詳細は cd-backend.yml の parameter-overrides で渡す部分を更新する PR を別途作成。
```

ステージング環境がある場合はそこで先に動作確認。

## Step B5: 本番 cutover (メンテナンスウィンドウ内、~30 分)

平日 22:00 JST のメンテナンスウィンドウで実施:

```bash
# 1. 利用者アナウンス（任意）
# 2. MariaDB を read-only にする
mariadb -h $MA_HOST -u admin -p"$MA_PASS" -e "SET GLOBAL read_only = ON;"

# 3. 最終差分の取り込み
# Step B3 の migrate-mariadb-to-postgres.sh を再実行（TRUNCATE→COPY なので再実行 OK）

# 4. ECS を新 DB に向けて再デプロイ
gh workflow run cd-backend.yml --ref main -R norman6464/FreStyle -f confirm=deploy

# 5. ヘルスチェック
curl -fsS https://api.normanblog.com/actuator/health
# {"status":"UP"} が返れば cutover 成功

# 6. しばらく動作確認 (15 分目安)
# 7. 問題なければアナウンス完了
```

## Step B6: 旧 MariaDB の停止 (1 週間後)

```bash
# 念のためスナップショット
aws rds create-db-snapshot --region ap-northeast-1 \
  --db-snapshot-identifier frestyle-prod-pre-postgres-cutover \
  --db-instance-identifier frestyle-prod

# CFn スタック削除
aws cloudformation delete-stack --region ap-northeast-1 --stack-name frestyle-prod-rds
```

## ロールバック計画

問題が発生した場合:

1. **直後 (Step B5 のヘルスチェック失敗)**:
   - ECS の DB_URL を MariaDB に戻し再デプロイ
   - MariaDB の `read_only = OFF` に戻す
   - Postgres スタックは残置（後日切戻す）

2. **数日後に問題発覚**:
   - MariaDB スナップショットから新規 DB 復元
   - 復元 DB に対して差分データ移行
   - 通常の cutover プロセスを再実行

## 検証チェックリスト

- [ ] ALB 経由で `/actuator/health` が UP
- [ ] `/api/auth/cognito/me` が 200
- [ ] フロントから AI セッション一覧取得 OK
- [ ] AI セッション作成・メッセージ送信 OK
- [ ] 練習シナリオ表示 OK
- [ ] スコア履歴表示 OK
- [ ] 通知一覧表示 OK
- [ ] WebSocket チャット OK
- [ ] CloudWatch Logs にエラーなし
- [ ] HikariCP のコネクションプールが正常

## 関連

- [DBMS 選定理由](../architecture/dbms-choice.md)
- [マルチテナントデータモデル](../architecture/multi-tenant-data-model.md)
- [Phase 0 マルチテナント基盤](./0001-multitenancy-foundation.md)
- frestyle-infrastructure: `docs/09-session-operations-runbook.md`
