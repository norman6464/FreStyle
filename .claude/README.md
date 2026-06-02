# .claude — Claude Code チーム設定

このディレクトリの **`settings.json` はリポジトリにコミットして共有する**チーム共通の Claude Code 設定です。
個人ごとの上書きは **`settings.local.json`（`.gitignore` 済 / 共有されない）** に書きます。

## 共有ポリシー（settings.json）

人的ミス・暴走を防ぐため、`bypassPermissions`（全許可）は**共有設定では使いません**。allow / ask / deny で制御します。

- **defaultMode: `acceptEdits`** — ファイル編集は自動承認（差分は PR でレビューする前提）。コマンドは下記に従う。
- **allow（無確認で実行可）** — ビルド・テスト・lint・フォーマット・ローカル git（`status` / `diff` / `add` / `commit` / `checkout` 等）・`gh pr create/view/list/comment`・`gitleaks` / `trivy` など安全な開発操作。
- **ask（毎回確認）** — **`git push` / `gh pr merge` / `gh pr review`**（リモート反映・マージ・承認）、`gh workflow run`（デプロイ起動）、`git rebase` / `git reset`。**既定では Claude に push / マージ させず、必ず人間が確認する**。
- **deny（常に禁止・ローカルでも上書き不可）** — `gh repo delete` / `gh secret` / `rm -rf`、`.env` 系の読み取り。致命的・機密操作は誰の環境でも禁止。

## 個人の例外（settings.local.json）

`deny` はローカルでも上書きできない絶対ルールですが、**`ask` はローカル設定で上書きできます**。
そのため、**リポジトリ管理者（norman6464）など信頼された個人は、自分の端末の `settings.local.json` で例外**にできます。

例（norman6464 の端末・`settings.local.json` / 共有されない）:

```json
{
  "permissions": { "defaultMode": "bypassPermissions" }
}
```

- これにより**その端末の Claude Code だけ** push / マージ等の確認がスキップされる（norman6464 は実質ルール対象外）。
- 一方で `rm -rf` / `.env` 読み取り等の **`deny` は維持**される（致命的操作は例外でも禁止）。
- 他のメンバーは `settings.local.json` を持たないので、共有ポリシー（push / merge は `ask`）が適用される。

## 運用

- 既定では Claude Code は「ローカルで変更・コミット・PR 作成」までを行い、**push とマージは人間が確認**する。
- ルールを変えたいときは `settings.json` を PR で変更する（チームで合意）。
- サーバ側の本当の歯止めは **GitHub のブランチ保護**（承認必須）と **production Environment の承認ゲート**。`.claude` 設定はクライアント側の補助ガード。
