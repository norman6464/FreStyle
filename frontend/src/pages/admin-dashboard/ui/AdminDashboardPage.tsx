
import { Link } from 'react-router-dom';
import { BuildingOffice2Icon } from '@heroicons/react/24/outline';

import Loading from '@/shared/ui/Loading';
import FormMessage from '@/shared/ui/FormMessage';
import PageIntro from '@/shared/ui/PageIntro';
import { useAdminDashboard } from '../model/useAdminDashboard';

/** 概況の数値カード。クリックで該当の管理画面へ遷移する。 */
function StatCard(props: {
  label: string;
  value: number;
  sub: string;
  to: string;
  icon: typeof BuildingOffice2Icon;
}) {
  const { label, value, sub, to, icon: Icon } = props;
  return (
    <Link
      to={to}
      className="block p-4 border rounded-lg bg-[var(--color-surface-1)] hover:bg-surface-2 transition-colors border-surface-3"
    >
      <div className="flex items-center gap-2 text-[var(--color-text-muted)]">
        <Icon className="w-4 h-4" />
        <span className="text-xs">{label}</span>
      </div>
      <p className="mt-1 text-2xl font-bold text-[var(--color-text-primary)]">
        {value}
      </p>
      <p className="text-xs text-[var(--color-text-muted)] mt-0.5">{sub}</p>
    </Link>
  );
}

/**
 * AdminDashboardPage — `/admin/dashboard`。super_admin 専用の運営概況。
 * 会社数（有効/無効）を一目で把握し、管理画面へ導く。
 * 全テナント横断の概況であり、通過条件はルート側の RequireRole が持つ。
 */
export default function AdminDashboardPage() {
  const { summary, loading, error } = useAdminDashboard();

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="運営ダッシュボード"
        description="全テナントの概況です。会社数を確認し、各管理画面へ移動できます。"
      />

      <FormMessage message={error ? { type: 'error', text: error } : null} />

      {loading ? (
        <Loading message="読み込み中..." className="min-h-[30vh]" />
      ) : summary ? (
        <div className="grid grid-cols-2 gap-3">
          <StatCard
            label="会社数"
            value={summary.companyTotal}
            sub={`有効 ${summary.companyActive} / 無効 ${summary.companyInactive}`}
            to="/admin/companies"
            icon={BuildingOffice2Icon}
          />
        </div>
      ) : null}
    </div>
  );
}
