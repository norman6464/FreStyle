<!--
PR タイトル・本文は日本語で。タイトルは Prefix を付ける（feat / fix / refactor / docs / test / chore / perf / style）。
詳細な規約は CONTRIBUTING.md を参照。
-->

## 概要

<!-- この PR で何を・なぜやったかを 1〜3 行で -->

## 変更内容

<!-- 主な変更点を箇条書きで -->
-

## テスト

<!-- どう検証したか（コマンド・結果）。新規コードには単体テストを付ける -->
-

## 関連 Issue

<!-- 例: Closes #123 / Refs #456 -->

---

### セルフチェック

- [ ] タイトルに Prefix（`feat` / `fix` / `refactor` / `docs` / `test` / `chore` 等）を付けた
- [ ] 新規・変更コードに単体テストを追加した（backend: `go test` / frontend: Vitest）
- [ ] backend を変更した場合 `make verify`（gofmt / vet / build / test / 3 linter）が通る
- [ ] handler / swaggo 注釈を変えた場合 `make openapi` を実行し `docs/` を commit した
- [ ] 仕様・手順を変えた場合は `docs/`（または該当 README）を更新した
- [ ] `main` への直接コミットではなく、PR 経由である
