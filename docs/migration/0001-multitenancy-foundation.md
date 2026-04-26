# Phase 0: マルチテナント基盤 + DBMS 切替（PostgreSQL）

## 目的

- 「会社」概念を導入し、既存データを `FreStyle` 社（id=1）にバックフィル
- 全ドメインテーブルに `company_id` を付与
- ロール (`super_admin` / `company_admin` / `trainee`) を導入
- DBMS を MariaDB → **PostgreSQL** に切替（Aurora 互換、RLS、JSONB を活用するため）

## ステップ概要

```
Step A. (アプリ無停止) 既存 MariaDB に最小限の変更を入れる
   A1. companies, company_signup_applications, invitations を新規作成
   A2. users に company_id, role を追加（NULL 許容）
   A3. FreStyle 社を INSERT、既存ユーザーを company_id=1, role='trainee' でバックフィル
   A4. 自分（河野拓真）を super_admin に昇格
   A5. 主要テーブルに company_id を追加（NULL 許容、既存行をバックフィル後 NOT NULL 化）

Step B. PostgreSQL 並行構築
   B1. CFn テンプレ更新 (rds.yml の Engine を postgres に)
   B2. 新スタック frestyle-prod-rds-postgres をデプロイ
   B3. AWS DMS で MariaDB → PostgreSQL に既存データ同期
   B4. DDL を PostgreSQL 形式で再構築（Step C のテーブル含む）
   B5. ステージングで Spring Boot 動作確認

Step C. PostgreSQL 専用機能
   C1. Row-Level Security 有効化
   C2. lessons (JSONB content) など Phase 1 用テーブル作成
   C3. Spring Boot のリクエストインターセプタで SET LOCAL app.current_company_id

Step D. 切替
   D1. 平日 22:00 JST メンテナンスウィンドウ通知
   D2. MariaDB を read-only、最終差分 DMS 同期
   D3. ECS Task Definition の DB_URL を PostgreSQL に切替
   D4. cd-backend を再実行
   D5. ヘルスチェック OK → アクセス再開
   D6. 1 週間経過後 MariaDB を停止 → 削除
```

## 即時実行（Step A 限定）

このマイグレーションは **既存 MariaDB に対して非破壊（追加のみ）** で実行できる。アプリは引き続き従来通り動作する。

実行ファイル: [`FreStyle/migrations/007_multitenancy_foundation.sql`](../../FreStyle/migrations/007_multitenancy_foundation.sql)

```bash
# EC2 踏み台 (test) 経由で RDS に接続
ssh ec2-test
mysql -h $RDS_HOST -u admin -p $DB_NAME < 007_multitenancy_foundation.sql

# 確認
mysql -h $RDS_HOST -u admin -p $DB_NAME -e "SELECT id, slug, name FROM companies;"
# 期待: 1 | frestyle | FreStyle 株式会社
mysql -h $RDS_HOST -u admin -p $DB_NAME -e "SELECT COUNT(*) FROM users WHERE company_id IS NULL;"
# 期待: 0
mysql -h $RDS_HOST -u admin -p $DB_NAME -e "SELECT id, username, email, role FROM users WHERE role='super_admin';"
# 期待: 河野拓真 1 行
```

## ロールバック (Step A)

`007_multitenancy_foundation.sql` は追加のみで既存カラム/データを破壊しないため、機能影響ゼロ。万一困った場合は:

```sql
ALTER TABLE practice_scenarios DROP COLUMN company_id;
-- ...他テーブルも同様
ALTER TABLE users DROP COLUMN company_id, DROP COLUMN role;
DROP TABLE invitations;
DROP TABLE company_signup_applications;
DROP TABLE companies;
```

## Step B-D の事前準備（ユーザー承認後に実施）

| 必要な準備 | 内容 |
|---|---|
| RDS PostgreSQL 用のサブネット / SG | 既存 SG を流用可能 |
| AWS DMS Replication Instance | 月数十ドル、移行完了後削除 |
| メンテナンスウィンドウの周知 | 平日 22:00-22:30 JST |
| ステージング環境 | 既存ステージングで本番フローを試験 |
| 切戻し条件の合意 | 「ヘルスチェック失敗 5 分以上 → 即時切戻し」 |

詳細スケジュールは別 PR で提案。

## 関連

- [DBMS 選定](../architecture/dbms-choice.md)
- [マルチテナントデータモデル](../architecture/multi-tenant-data-model.md)
