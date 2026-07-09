# ホームダッシュボード（MenuPage）のレイアウトと表示タイミング

`/`（ホーム＝[MenuPage](../frontend/src/pages/MenuPage.tsx)）のレイアウトと、メニューカード／
パーソナライズ統計（右サイドバー）の表示タイミングについて。

## 構成

2 カラム構成（`lg` 以上で横並び、それ未満で縦積み）。

- **左カラム**: ロール別のメニューカード群（[FeatureCard](../frontend/src/pages/MenuPage.tsx)）
  - super_admin … 管理機能のみ
  - company_admin … 学習 + ツール + 管理
  - trainee … 学習 + ツール（AI カードは `aiChatEnabledForTrainees` に従う）
- **右サイドバー**（super_admin 以外）: 学習統計
  [DashboardStats](../frontend/src/components/dashboard/DashboardStats.tsx)
  - KPI（連続学習 / 演習数 / 正答率 / 章完了）2×2
  - 学習カレンダー（過去 90 日のヒートマップ）
    [LearningCalendar](../frontend/src/components/dashboard/LearningCalendar.tsx)

メニューカードのグリッドは `sm:grid-cols-2`（最大 2 列）。以前は `lg:grid-cols-3` だったが、
1 セクション 2 カードのとき 3 列目が空いて間延びして見えたため 2 列に統一した。

## 表示タイミング（メニューと統計を同時に出す）

学習者向けダッシュボードでは、**メニューカードと右サイドバー統計を同時に表示**する。

- 統計は `GET /api/v2/me/dashboard`（[useUserDashboard](../frontend/src/hooks/useUserDashboard.ts)）で
  非同期取得する。以前はメニュー（データ非依存）が先に描画され、統計が遅れて差し込まれていたため
  「メニューだけ先に出てサイドバーが後からポップインする」ちらつきが発生していた。
- これを防ぐため、**統計ロード中は左右ともスケルトンを表示**し、ロード完了で両カラムを同時に実カードへ
  差し替える。判定は `waitingForStats = !isSuperAdmin && loading`。
- ウェルカム見出し（データ非依存）は即時表示する。
- super_admin は統計を取得しないため（`enabled: false`）即時にメニューを表示する。
- 統計取得に失敗（`error`）した場合はロード完了扱いとし、メニューのみ表示してページを使える状態に保つ
  （サイドバーは出さない）。

スケルトンは実レイアウトと寸法を揃え（カード高さ・KPI グリッド・カレンダー領域）、
差し替え時のレイアウトシフトを抑える。

## 学習カレンダー（ヒートマップ）

過去 90 日の活動量（exercise + lesson + ai + note の合計）を 4 段階の色で表示する。
セルは **曜日（行）× 週（列）のグリッド**（`grid-flow-col grid-rows-7`）に並べ、先頭の曜日ぶんを
空セルで詰めて各列が日曜始まりの 1 週間に揃うようにしている。以前は `flex-wrap` で折り返していたが、
週の境界が揃わず崩れて見えたためグリッドに変更した。

## ロール別サイドバー（FRESTYLE-103）

ホーム右サイドバーの中身はロールで出し分ける:

| ロール | サイドバー | データ源 |
|---|---|---|
| trainee | 自分の学習統計（`DashboardStats`。従来どおり） | `GET /api/v2/me/dashboard` |
| company_admin | **メンバーの学習状況**（`CompanyLearningPanel`） | `GET /api/v2/admin/members/learning-summary` |
| super_admin | 非表示（従来どおり） | — |

company_admin はメンター（自社メンバーの学習支援・管理）が主用途のため、
自分の学習統計ではなく自社 trainee の学習状況サマリーを表示する。
自分用の `/me/dashboard` は取得しない。

### メンバーの学習状況サマリーの中身

- 在籍メンバー数（論理削除済みを除く自社 trainee）
- 今日学習した人数 / 直近 7 日（今日を含む）で学習した人数
- 直近アクティブメンバー最大 5 名（名前・最終学習日・直近 7 日の活動回数）
- 従業員一覧（/admin/members）へのリンク

集計は backend の `GetCompanyLearningSummaryUseCase` +
`CompanyLearningActivitySummarizer`（users × user_daily_activities の
LEFT JOIN・GROUP BY 1 クエリ、N+1 回避）。日付境界は `/me/dashboard` と同じ UTC。
「活動」の定義も従来どおり user_daily_activities（演習提出・章完了・AI チャット・ノート作成）。
