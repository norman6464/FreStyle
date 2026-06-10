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

## ディレクトリ規約（年フォルダ + メタデータ索引）

Design Doc は **年フォルダ** に置き、**ファイルは動かしません**（追記のみ＝リンク切れ・履歴の混乱を防ぐ）。
領域 / ステータス / 日付といった「別の軸」は、各 doc 冒頭の **frontmatter** と、この README の
**索引表**で引きます。フォルダは 1 軸しか表現できないため、多軸はメタデータ側で持つ方針です。

```
docs/design/
├── README.md          ← この説明 + 索引表
├── _TEMPLATE.md       ← 蒸留用テンプレート（frontmatter 付き。コピーして使う）
├── _scripts/
│   └── build_index.py ← frontmatter を集計して README の索引表を再生成
└── 2026年/
    ├── 0001-<feature-slug>.md
    └── ...
```

- フォルダ名は **`<西暦>年`**（例: `2026年`）。起票・蒸留した年で分ける
- ファイル名は **`000N-<feature-slug>.md`**（年ごとに 0001 から連番、英小文字 kebab の slug）
- 各 doc 冒頭に **frontmatter** を付ける（`_TEMPLATE.md` 参照）:

  ```yaml
  ---
  status: approved     # draft / reviewing / approved / superseded / rejected
  area: backend        # auth / backend / frontend / infra / data / docs
  date: 2026-06-09
  issue: norman6464/FreStyle#1846   # 対応 Issue があれば
  # supersedes: 0001   # 置き換える旧 doc があれば
  ---
  ```

### 索引（自動生成）

doc を追加・更新したら **`python3 docs/design/_scripts/build_index.py`** を実行して下表を再生成する
（下のマーカー行で囲まれた範囲が置き換わる）。領域・ステータス・日付は表のソートや `grep` で引ける。

<!-- INDEX:START -->
| # | タイトル | 領域 | ステータス | 日付 |
|---|---|---|---|---|
| [0001](2026年/0001-trainee-ai-agent-toggle.md) | trainee 向け AI エージェント機能の利用可否トグル | backend | approved | 2026-06-08 |
| [0002](2026年/0002-backend-java-to-go-revert.md) | バックエンドを Java(Spring Boot) から Go へ差し戻す | backend | approved | 2026-06-09 |
| [0003](2026年/0003-brand-color-buttons.md) | ブランドカラー（青）でアクションボタンを統一 | frontend | approved | 2026-06-09 |
| [0004](2026年/0004-code-runner-separation.md) | コード実行サンドボックスを backend から分離してイメージを軽量化 | infra | reviewing | 2026-06-10 |
<!-- INDEX:END -->

## テンプレート

蒸留版の雛形は [`_TEMPLATE.md`](./_TEMPLATE.md)。Issue 起票用の雛形は
[`.github/ISSUE_TEMPLATE/design-doc.md`](../../.github/ISSUE_TEMPLATE/design-doc.md)。両者は同じ 7 セクション構成です。
