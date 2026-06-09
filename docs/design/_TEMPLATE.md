<!--
蒸留用テンプレート。docs/design/<年>/000N-<feature-slug>.md にコピーして使う。
Design Doc Issue が承認(status:approved)されたら、確定した設計をここに残す。
冒頭の frontmatter は README の索引表に集計される（_scripts/build_index.py で再生成）。
-->
---
status: draft        # draft / reviewing / approved / superseded / rejected
area: backend        # auth / backend / frontend / infra / data / docs
date: YYYY-MM-DD     # 蒸留した日
issue: norman6464/FreStyle#NN   # 対応 Issue があれば（無ければこの行を削除）
# supersedes: 000N   # 置き換える旧 doc があれば記載
---

# 000N: <タイトル>

## 1. 基本情報

| 項目 | 内容 |
|---|---|
| タイトル | <機能名 / 変更名> |
| 担当者 | @ |
| レビュアー | @ |
| ステークホルダー | <関係者> |
| ステータス | Approved |
| 関連 | Issue #NN / PR #NN |

## 2. 背景と目的 (Background & Objective)

**背景**: <なぜ必要か。課題やビジネス上の理由>

**目的**: <何を達成したいか>

## 3. スコープ (Goals / Non-Goals)

**Goals**
- <スコープに入る範囲>

**Non-Goals**
- <スコープから外すもの>

## 4. アーキテクチャ・設計詳細 (Architecture & Design)

### 概要設計
<構成図 / シーケンス図>

### データ構造 / API 仕様
<テーブル定義(V{n}) / API インターフェース>

### 変更点
<既存から何が変わるか。touch する層>

## 5. 検討した代替案 (Alternatives & Trade-offs)

<却下した代替案と Pros/Cons。設計判断の理由>

## 6. 運用・非機能要件 (Operations & Non-Functional)

- 監視・ログ:
- セキュリティ:
- パフォーマンス:
- 移行・リリース計画（ロールバック含む）:

## 7. テストプラン (Testing)

- ユニット:
- 結合:
- 手動・本番検証:
