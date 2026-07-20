import { useSelector } from 'react-redux';
import { Navigate, Link } from 'react-router-dom';
import { BuildingOffice2Icon, InboxArrowDownIcon } from '@heroicons/react/24/outline';
import type { RootState } from '@/store';
import Loading from '@/components/Loading';
import PageIntro from '@/shared/ui/PageIntro';
import { useAdminDashboard } from '@/hooks/useAdminDashboard';

/** 概況の数値カード。クリックで該当の管理画面へ遷移する。 */
function StatCard(props: {
  label: string;
  value: number;
  sub: string;
  to: string;
  icon: typeof BuildingOffice2Icon;
  emphasize?: boolean;
}) {
  const { label, value, sub, to, icon: Icon, emphasize } = props;
  return (
    <Link
      to={to}
      className={`block p-4 border rounded-lg bg-[var(--color-surface-1)] hover:bg-surface-2 transition-colors ${
        emphasize ? 'border-amber-300' : 'border-surface-3'
      }`}
    >
      <div className="flex items-center gap-2 text-[var(--color-text-muted)]">
        <Icon className="w-4 h-4" />
        <span className="text-xs">{label}</span>
      </div>
      <p className={`mt-1 text-2xl font-bold ${emphasize ? 'text-amber-700' : 'text-[var(--color-text-primary)]'}`}>
        {value}
      </p>
      <p className="text-xs text-[var(--color-text-muted)] mt-0.5">{sub}</p>
    </Link>
  );
}

/**
 * AdminDashboardPage — `/admin/dashboard`。super_admin 専用の運営概況。
 * 会社数（有効/無効）と承認待ちの利用申請件数を一目で把握し、各管理画面へ導く。
 */
export default function AdminDashboardPage() {
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);
  const authLoading = useSelector((state: RootState) => state.auth.loading);
  const role = useSelector((state: RootState) => state.auth.role);
  // super_admin のときだけ admin API を取得する（リダイレクト対象に権限外アクセスを試行させない）。
  const canView = isAdmin && role === 'super_admin';
  const { summary, loading, error } = useAdminDashboard(canView);

  if (authLoading) return <Loading message="認証情報を確認中..." className="min-h-[50vh]" />;
  // 運営ダッシュボードは全テナント横断の概況なので super_admin 専用。
  if (!canView) return <Navigate to="/" replace />;

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="運営ダッシュボード"
        description="全テナントの概況です。会社数と承認待ちの利用申請を確認し、各管理画面へ移動できます。"
      />

      {error && (
        <div role="alert" className="p-3 rounded border border-rose-300 bg-rose-50 text-rose-800 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <Loading message="読み込み中..." className="min-h-[30vh]" />
      ) : summary ? (
        <>
          <div className="grid grid-cols-2 gap-3">
            <StatCard
              label="会社数"
              value={summary.companyTotal}
              sub={`有効 ${summary.companyActive} / 無効 ${summary.companyInactive}`}
              to="/admin/companies"
              icon={BuildingOffice2Icon}
            />
            <StatCard
              label="承認待ちの申請"
              value={summary.pendingApplications}
              sub={`全 ${summary.applicationTotal} 件`}
              to="/admin/applications"
              icon={InboxArrowDownIcon}
              emphasize={summary.pendingApplications > 0}
            />
          </div>

          <section className="space-y-2">
            <h2 className="text-sm font-semibold text-[var(--color-text-secondary)]">承認待ちの利用申請</h2>
            {summary.recentPending.length === 0 ? (
              <p className="text-sm text-[var(--color-text-muted)]">承認待ちの申請はありません。</p>
            ) : (
              <ul className="space-y-2">
                {summary.recentPending.map((app) => (
                  <li
                    key={app.id}
                    className="p-3 border border-surface-3 rounded-lg bg-[var(--color-surface-1)] flex items-center justify-between gap-3"
                  >
                    <div className="min-w-0">
                      <p className="text-sm font-medium truncate">{app.companyName}</p>
                      <p className="text-xs text-[var(--color-text-muted)] truncate">
                        {app.applicantName}（{app.email}）
                      </p>
                    </div>
                    <Link
                      to="/admin/applications"
                      className="text-xs px-3 py-1.5 border border-surface-3 rounded text-[var(--color-text-secondary)] hover:bg-surface-2 transition-colors flex-shrink-0"
                    >
                      確認
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </>
      ) : null}
    </div>
  );
}
