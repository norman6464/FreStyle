import { useState } from 'react';
import { ArrowPathIcon, ClockIcon } from '@heroicons/react/24/outline';

interface MaintenancePageProps {
  /** 「再試行」ボタンが押されたときに呼ぶ。useBackendHealth.recheck を渡す。 */
  onRetry: () => void;
}

/**
 * バックエンド停止時のメンテナンスページ。
 *
 * 表示条件: useBackendHealth が `'unhealthy'` のとき App.tsx が本ページを描画する。
 * 主な発生源:
 *   - 毎日 22:00 JST の scheduled-stop で ECS / ALB / RDS が destroy される時間帯
 *   - 突発的な ECS 障害 / ALB 設定ミス / DNS 解決失敗
 *
 * 自動復帰: 親 (useBackendHealth) が 15 秒間隔で health を叩き続け、応答が
 * 返るようになった瞬間に App 側で `<Routes>` 表示に戻る。手動「再試行」ボタンは
 * その poll を待たずに即座に再チェックする。
 */
export default function MaintenancePage({ onRetry }: MaintenancePageProps) {
  // 「再試行中…」を 1.5 秒程度表示してフィードバックを与える
  const [retrying, setRetrying] = useState(false);

  // サポート連絡先は環境変数で注入。未設定時は連絡先行を出さない（プレースホルダ表示で混乱させないため）。
  const supportEmail = (import.meta.env.VITE_SUPPORT_EMAIL ?? '').trim();

  const handleRetry = () => {
    setRetrying(true);
    onRetry();
    window.setTimeout(() => setRetrying(false), 1500);
  };

  // バックエンドが停止する時間帯 (JST 22:00〜翌 07:00) を表示。本サービスは scheduled-stop で
  // 毎日この帯に止まることがアナウンスされている前提のため、ユーザーに「いつ戻るか」の
  // 期待値を示すことで体験を悪くしないようにする。
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
        <p className="text-sm text-[var(--color-text-muted)] leading-relaxed mb-6">
          サーバが一時的にアクセスできない状態です。
          <br />
          自動的に再接続を試みていますので、しばらくお待ちください。
        </p>

        <div className="flex items-start gap-3 px-4 py-3 rounded-lg border border-[var(--color-surface-3)] bg-[var(--color-surface-2)] text-left mb-6">
          <ClockIcon className="w-5 h-5 text-[var(--color-text-muted)] flex-shrink-0 mt-0.5" />
          <div className="text-xs text-[var(--color-text-muted)] leading-relaxed">
            <p className="font-medium text-[var(--color-text-secondary)] mb-1">
              定期メンテナンス時間帯
            </p>
            <p>
              料金節約のため、毎日 <strong>22:00〜翌 07:00 JST</strong> はサービスを停止しています。
              この時間帯の利用はできません。
            </p>
          </div>
        </div>

        <button
          type="button"
          onClick={handleRetry}
          disabled={retrying}
          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-500 text-white text-sm font-medium hover:bg-primary-600 transition-colors disabled:opacity-60 disabled:cursor-wait"
        >
          <ArrowPathIcon
            className={`w-4 h-4 ${retrying ? 'animate-spin' : ''}`}
            aria-hidden="true"
          />
          {retrying ? '再接続中...' : '再試行'}
        </button>

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

