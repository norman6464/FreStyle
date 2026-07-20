/**
 * バックエンド停止時のメンテナンスページ。
 *
 * 表示条件: useBackendHealth が `'unhealthy'` のとき App.tsx が本ページを描画する。
 * 主な発生源:
 *   - cd-backend デプロイ中の rolling update 谷間 (= 数秒〜1 分)
 *   - 突発的な ECS 障害 / ALB 設定ミス / DNS 解決失敗
 *   - Supabase 側のメンテナンス / 接続エラー
 *
 * 経緯: 2026-05 まで「毎日 22:00 JST の scheduled-stop で ECS / ALB / RDS を
 * destroy する時間帯」が発生源だったが、 Supabase 移行 (PR #1730) で
 * scheduled-stop / scheduled-start ワークフローを廃止。 現在は 24/7 稼働。
 *
 * 自動復帰: 親 (useBackendHealth) が 15 秒間隔で health を叩き続け、応答が
 * 返るようになった瞬間に App 側で `<Routes>` 表示に戻る。
 * 「再試行」ボタン / 「定期メンテナンス時間帯」 案内カードは ユーザ要望で 削除済。
 */
export default function MaintenancePage() {
  // サポート連絡先は環境変数で注入。未設定時は連絡先行を出さない（プレースホルダ表示で混乱させないため）。
  const supportEmail = (import.meta.env.VITE_SUPPORT_EMAIL ?? '').trim();

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-surface-1)] px-6">
      <div className="max-w-lg w-full text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 mb-6">
          <img
            src="/favicon.svg"
            alt="FreStyle"
            className="w-12 h-12"
            width={48}
            height={48}
          />
        </div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)] mb-3">
          ただいまメンテナンス中です
        </h1>
        <p className="text-sm text-[var(--color-text-muted)] leading-relaxed">
          サーバーにアクセスできない状態です。
          <br />
          自動的に再接続を試みていますので、しばらくお待ちください。
        </p>

        {supportEmail && (
          <p className="mt-8 text-xs text-[var(--color-text-faint)]">
            状況が長く続く場合はサポート（
            <a
              href={`mailto:${supportEmail}`}
              className="underline hover:text-[var(--color-text-muted)]"
            >
              {supportEmail}
            </a>
            ）までご連絡ください。
          </p>
        )}
      </div>
    </div>
  );
}
