# Design Docs（設計ドキュメント）

FreStyle の「実装前に設計を共有し、レビュー・承認を得る」ための Design Doc を置く場所です。

## いつ書くか

- **Design Doc を書く**: 大きめの機能追加 / アーキテクチャ変更 / データモデル変更 / 認証・権限に関わる変更 / 複数層をまたぐ変更
- **書かなくてよい**: 軽微な修正・typo・見た目だけの変更（`task` Issue で十分）

## フロー（ハイブリッド運用）

設計の「議論・承認」は **GitHub Issue**、確定した設計判断の「恒久保管」は **`docs/design/`** に役割分担します。

```
1. 起票    : New Issue → "Design Doc / 設計ドキュメント" テンプレートで起票
             (ラベル design-doc / status:draft が自動付与)
2. レビュー : ラベルを status:draft → status:reviewing に。コメントで議論
3. 承認    : status:approved に。GitHub Projects #3 のカンバンで進捗管理
4. 蒸留    : 承認された確定設計を docs/design/<年>/000N-<feature>.md に蒸留して残す
             (実装 PR と同じ or 直後の PR で。CLAUDE.md §7「作ったものは docs に残す」)
5. 実装    : 1 機能 = 1 PR。Design Doc Issue を close するのは実装 PR のマージ時
```

- **Issue** = 提案・レビュー・承認の「生もの」（議論の履歴が残る）
- **`docs/design/`** = 承認後の確定設計の「保管庫」（バージョン管理され、半年後の参照に耐える）

### PR は必ず Issue に紐付ける

実装 PR は、対応する Design Doc Issue（または task Issue）に**必ずリンク**する。PR 本文に
`Closes #NN`（マージで自動 close）または `Related to #NN`（参照のみ）を書く。

- レビュアーが「この PR は何の設計に基づくか」を Issue の Design Doc から辿れる
- カンバン（GitHub Projects #3）上で Issue ↔ PR が連動し、進捗が一目で分かる

## ディレクトリ規約（年フォルダ）

Design Doc は **年フォルダ** で整理します。

```
docs/design/
├── README.md          ← この説明
├── _TEMPLATE.md       ← 蒸留用の Markdown テンプレート（コピーして使う）
└── 2026年/
    ├── 0001-<feature-slug>.md
    ├── 0002-<feature-slug>.md
    └── ...
```

- フォルダ名は **`<西暦>年`**（例: `2026年`）。起票・蒸留した年で分ける
- ファイル名は **`000N-<feature-slug>.md`**（年ごとに 0001 から連番、英小文字 kebab の slug）
- 対応する Design Doc Issue があれば、doc 冒頭の「関連」に Issue 番号を書いてリンクする

## テンプレート

蒸留版の雛形は [`_TEMPLATE.md`](./_TEMPLATE.md)。Issue 起票用の雛形は
[`.github/ISSUE_TEMPLATE/design-doc.md`](../../.github/ISSUE_TEMPLATE/design-doc.md)。両者は同じ 7 セクション構成です。
