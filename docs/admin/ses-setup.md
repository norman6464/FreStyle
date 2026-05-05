# SES セットアップ手順（招待マジックリンク用）

招待マジックリンクメール（[invitation-magic-link-flow.md](./invitation-magic-link-flow.md)）を実環境で動かすために、AWS Console / IAM 側で必要な手動作業をまとめる。

PR-B はコード側のみ実装しており、以下のインフラ作業を行うまでは `SES_FROM_ADDRESS` 未設定のままフォールバックモード（ログにトークン出力）で動く。

---

## 1. 送信ドメインの検証（DKIM）

ドメイン全体を検証するほうが運用が楽（メールアドレスごとの個別検証は不要になる）。

1. AWS Console → **Amazon SES** → **Configuration** → **Identities** → **Create identity**
2. **Identity type**: `Domain` を選択
3. **Domain**: `normanblog.com` を入力
4. **Use a custom MAIL FROM domain**: 任意（推奨。`mail.normanblog.com` など）
5. **Easy DKIM**: 有効化（RSA 2048 bit を選択）
6. 「Create identity」をクリック → DNS レコード（CNAME 3 件）が表示される
7. Route 53（`normanblog.com` のホストゾーン）に **DKIM の CNAME 3 件** を追加
   - `<token1>._domainkey.normanblog.com → <token1>.dkim.amazonses.com`
   - 同様に token2 / token3
8. 数分〜30 分後、SES Identities ページで `Verification status: Verified` / `DKIM: Successful` になることを確認

検証が通ると、このドメイン配下の任意のローカルパート（`noreply@normanblog.com` など）から送信できるようになる。

---

## 2. From アドレスの設計

- 推奨: `noreply@normanblog.com`
- 表示名付きで `FreStyle <noreply@normanblog.com>` の形式を `SES_FROM_ADDRESS` に入れる
- 返信を想定しないことを明示するため `noreply@` を採用

---

## 3. SES サンドボックス中の宛先検証

リリース直後はサンドボックスのままで運用する（送信上限 200 通/日 で十分）。

サンドボックス中は **宛先メールアドレスも SES で検証済** である必要がある:

1. SES → Identities → **Create identity** → `Email address`
2. 招待先のメールアドレス（例: 社内テスト用アドレス）を 1 件ずつ追加
3. 該当アドレスに「Verify」メールが届くのでクリックしてもらう

サンドボックス解除（=`Verified` 不要で誰にでも送信可）は、社内ベータ運用の手応えを見てから AWS Console → SES → Account dashboard → Request production access で申請する。

---

## 4. ECS Task Role に `ses:SendEmail` を追加

タスクロール（`taskRoleArn` のロール、ECR や DynamoDB 用にも使われている）に SES 送信権限を加える。

1. IAM Console → Roles → `ecsTaskExecutionRole`（または現行のタスクロール）を選択
2. **Add permissions** → **Create inline policy**
3. JSON タブで以下を貼り付け:

   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "ses:SendEmail",
           "ses:SendRawEmail"
         ],
         "Resource": [
           "arn:aws:ses:ap-northeast-1:010928196665:identity/normanblog.com",
           "arn:aws:ses:ap-northeast-1:010928196665:identity/*@normanblog.com"
         ]
       }
     ]
   }
   ```

4. 名前: `ses-send-invitation` などで保存

`Resource` を ID に絞ることで、誤って他ドメインから送信できないようにする。

---

## 5. ECS タスク定義に環境変数を追加

`task-definition.json` の `environment` に以下を追加（または GitHub Actions の deploy ワークフロー / SSM Parameter Store 経由で注入）:

```json
{
  "name": "SES_REGION",
  "value": "ap-northeast-1"
},
{
  "name": "SES_FROM_ADDRESS",
  "value": "FreStyle <noreply@normanblog.com>"
},
{
  "name": "APP_BASE_URL",
  "value": "https://normanblog.com"
}
```

`COGNITO_USER_POOL_ID` は PR-B で不要になったので削除しても良い（残しても無視される）。

---

## 6. 動作確認手順

1. ECS にデプロイ後、CloudWatch Logs `/ecs/fre_style_ecs_task` で `SES client init` のログが出ていないこと（成功時はログ無し）を確認
2. `/api/v2/admin/invitations` に管理者として POST で招待を作る
3. SES Console → **Sending statistics** で送信カウンタが 1 増えていることを確認
4. 受信側のメールクライアントで「FreStyle へようこそ」が届く
5. メール本文中の `https://normanblog.com/invitations/accept?token=...` を踏むと、PR-D 実装後はフロントの受諾画面に遷移する（PR-C / PR-D 未着手の現時点では 404）

---

## 7. トラブルシューティング

| 症状 | 原因 | 対処 |
|---|---|---|
| `MessageRejected: Email address is not verified` | サンドボックス中で宛先が未検証 | 宛先を SES Identities に追加し検証メール承認 |
| `AccessDenied: not authorized to perform ses:SendEmail` | タスクロールに権限なし | 上記 4. のインライン policy を確認 |
| バックエンドログに `WARN: SES client init failed` | リージョン誤り / 認証情報 | `SES_REGION` の typo・タスクロール ARN を再確認 |
| ログに `token=...` だけ出てメールが来ない | `SES_FROM_ADDRESS` または `APP_BASE_URL` が空 | task definition に env が注入されているか確認 |

---

## 8. 後で考えること

- バウンス / 苦情通知（SNS topic 経由）→ 招待を `canceled` に自動マークするバッチ
- 専用 IAM ユーザー（タスクロールではなく）への切り替え（最小権限の原則）
- DMARC / SPF レコードの整備（DKIM だけでも動くが、配信率向上のため）
