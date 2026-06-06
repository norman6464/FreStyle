---
name: 開発タスク / 機能移植
about: 機能追加・Java 移行などの開発タスクを起票する（1 タスク = 1 PR）
title: 'feat: '
labels: task
---

## 目的・背景

<!-- なぜこれをやるか。解決したい課題・狙い -->

## 完了条件（Definition of Done）

<!-- これが満たされたら完了、という条件を箇条書きで -->

- [ ]
- [ ]

## 想定する変更

<!-- 触る範囲: backend-java(controller / service / repository / entity / dto) / frontend / infra / docs -->
<!-- 追加・変更する API パス、画面、マイグレーション(V{n}) など -->

## 進行（カンバンのステータスに対応）

<!-- ボードの「ステータス」を進めつつ、ここもチェックしていく -->

- [ ] **開発中**: 実装 + テスト（`./gradlew build` / `vitest`）
- [ ] **レビュー**: PR 作成 → CodeRabbit 対応 → squash merge
- [ ] **ステージング検証**: ローカル / ステージングで動作確認
- [ ] **リリース**: 本番デプロイ（backend = ECS force-new-deployment / frontend = CD）
- [ ] **リリース検証**: 本番で health + 当該エンドポイント・画面を確認

## 補足

<!-- 設計メモ・代替案・関連 Issue / PR / docs -->
