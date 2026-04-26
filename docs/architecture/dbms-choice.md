# DBMS 選定: MariaDB → PostgreSQL への移行

## 結論

**PostgreSQL（RDS db.t4g.micro → 将来 Aurora Serverless v2）** に移行する。

## 比較表

| 項目 | MariaDB（現状） | PostgreSQL（推奨） |
|---|---|---|
| **マルチテナント Row-Level Security** | アプリ層フィルタのみ | **DB 層 RLS あり**（防御 2 層化） |
| **JSON 操作** | JSON 関数あり、インデックスは虚弱 | **JSONB + GIN index** で実用的 |
| **全文検索** | FULLTEXT INDEX（性能限定） | **tsvector + GIN**、後で pgvector で AI 検索拡張可 |
| **CTE / Recursive クエリ** | あり（一部制約） | あり（標準準拠） |
| **Window 関数** | あり | あり（より豊富） |
| **拡張機能** | 少ない | 豊富（pgvector / pg_trgm / pg_partman / TimescaleDB / pgaudit） |
| **AWS RDS サポート** | あり | あり |
| **AWS Aurora 互換** | Aurora MySQL | **Aurora PostgreSQL**（Serverless v2 で 0.5 ACU 起動） |
| **Spring Data JPA / Hibernate** | 互換 | 互換（Dialect 切替のみ） |
| **OSS コミュニティ** | やや限定 | 圧倒的に大きい（B2B SaaS 標準） |
| **コスト (db.t4g.micro 24/7)** | ~$13/月 | ~$13/月（同水準） |

## 移行を選ぶ理由（B2B SaaS の文脈で）

### 1. **Row-Level Security (RLS) でテナント漏洩事故を防ぐ**

PostgreSQL なら DB 側で次のようなポリシーが書ける:

```sql
ALTER TABLE courses ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON courses
  USING (company_id = current_setting('app.current_company_id')::bigint);
```

アプリ層で `WHERE company_id = ?` を 1 箇所書き忘れても、DB 側が拒絶する。**B2B SaaS で最重要の "テナントデータ漏洩" リスクを構造的に下げられる**。MariaDB には同等機能がない。

### 2. **JSONB で教材コンテンツを柔軟に格納**

レッスンの種類（Markdown / クイズ / コーディング / 写経）は構造が異なる。JSONB なら:

```sql
-- クイズの選択肢を 1 列に格納し、その場でクエリ
SELECT id FROM lessons WHERE content @> '{"type": "quiz"}'::jsonb;
```

スキーマ追加なしに新レッスン種別を追加可能。MariaDB の JSON 型はこのレベルでは厳しい。

### 3. **将来の AI 統合に備えた pgvector**

教材検索を「キーワード一致」から「意味検索（embedding）」に進化させる際、PostgreSQL は **pgvector 拡張**で対応可能:

```sql
CREATE EXTENSION vector;
ALTER TABLE lessons ADD COLUMN embedding vector(1536);
```

Bedrock Titan Embeddings と組み合わせると "新卒が質問した文章に意味的に近い教材を提案" が現実的。MariaDB では別 DB が必要になる。

### 4. **Aurora PostgreSQL Serverless v2 への移行パスがクリーン**

事業が伸びたら Aurora Serverless v2 に切替えれば自動スケール。MariaDB は Aurora MySQL になるが、PostgreSQL の方がエンタープライズ顧客との親和性が高い。

### 5. **今が一番安い乗り換えタイミング**

- 現状テーブル数: ~25
- 既存ユーザー数: 数名（ベータ）
- 既存教材データ: 静的シナリオ + テンプレート程度
- → 今やれば移行 SQL は数百行で済む。100 社入った後では桁違いの工数になる

## 移行戦略

### Phase 0a: 並行運用（破壊的変更ゼロ）

1. RDS PostgreSQL インスタンスを **新規作成**（既存 MariaDB はそのまま稼働）
2. DDL を PostgreSQL 形式で書き直し（後述のスキーマ設計に従う）
3. AWS Database Migration Service (DMS) または `pg_loader` で MariaDB → PostgreSQL に既存データを 1 回コピー
4. Spring Boot の `application.properties` を切替できるよう環境変数化（既に対応済み: `DB_URL`）
5. ステージング環境で動作確認

### Phase 0b: 切替（メンテナンスウィンドウ）

1. 平日 22:00（既存の scheduled-stop タイミング）にユーザー通知済みでメンテイン
2. MariaDB を read-only に → 最終差分を DMS で同期
3. ECS の DB_URL を PostgreSQL に切替 → 再デプロイ
4. ヘルスチェック OK → アクセス再開
5. 24h 様子を見て問題なければ MariaDB インスタンス停止（1 週間後に削除）

### Phase 0c: PostgreSQL 専用機能の段階的採用

- マルチテナント RLS の有効化（Phase 1 開始時）
- JSONB レッスンコンテンツ（Phase 1 教材実装時）
- pgvector（Phase 2 以降）

## CFn / IaC の変更

`frestyle-infrastructure/infrastructure/cloudformation/templates/runtime/rds.yml` を以下のように改修:

```yaml
DBInstance:
  Type: AWS::RDS::DBInstance
  Properties:
    Engine: postgres                    # mariadb → postgres
    EngineVersion: "16.6"               # PostgreSQL 16 系
    DBInstanceClass: db.t4g.micro
    AllocatedStorage: 20
    StorageType: gp3
    DeletionProtection: true
    BackupRetentionPeriod: 7
    PubliclyAccessible: false
    # ... (他の設定はそのまま)
```

別スタック名 `frestyle-prod-rds-postgres` で先に作り、検証後に DNS / 接続を切替える。旧 `frestyle-prod-rds` (MariaDB) は事後に destroy。

## アプリケーション側の変更

| 場所 | 変更点 |
|---|---|
| `FreStyle/build.gradle` | `mysql-connector-j` → `org.postgresql:postgresql` |
| `application.properties` | `spring.jpa.properties.hibernate.dialect` を `MariaDBDialect` → `PostgreSQLDialect` |
| `application.properties` | `spring.datasource.driver-class-name` を `com.mysql.cj.jdbc.Driver` → `org.postgresql.Driver` |
| `schema.sql` | MariaDB 構文 → PostgreSQL 構文（CHARSET / COLLATE 削除、AUTO_INCREMENT → BIGSERIAL/IDENTITY） |
| マイグレーション運用 | `schema.sql` の手書き → **Flyway** 導入（バージョン管理されたマイグレーション） |

## ロールバック計画

万一問題が発生した場合:

1. ECS の `DB_URL` を MariaDB に戻して再デプロイ → 即座に旧環境復帰
2. PostgreSQL 側で発生したデータは差分同期で MariaDB に戻すか諦める
3. 原因調査 → 修正 → 再度切替

MariaDB インスタンスは **切替後も最低 1 週間は停止せず保持**することで安全側に倒す。

## 参考リンク

- [PostgreSQL Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [pgvector](https://github.com/pgvector/pgvector)
- [AWS DMS for Heterogeneous Migration](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.html)
- [Aurora Serverless v2](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2.html)
