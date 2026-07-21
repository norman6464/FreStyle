
import { Navigate, Link } from 'react-router-dom';
import { useAppSelector } from '@/shared/lib/store';
import { BuildingOffice2Icon, CheckIcon, XMarkIcon, UserPlusIcon } from '@heroicons/react/24/outline';

import Loading from '@/shared/ui/Loading';
import PageIntro from '@/shared/ui/PageIntro';
import { useToast } from '@/shared/lib/hooks/useToast';
import { useCompanyApplications } from '../model/useCompanyApplications';
import type {
  CompanyApplication,
  CompanyApplicationStatus,
} from '@/entities/company';

const STATUS_LABEL: Record<CompanyApplicationStatus, string> = {
  pending: '承認待ち',
  approved: '承認済み',
  rejected: '却下',
};

const STATUS_CLASS: Record<CompanyApplicationStatus, string> = {
  pending: 'bg-amber-100 text-amber-800',
  approved: 'bg-emerald-100 text-emerald-700',
  rejected: 'bg-rose-100 text-rose-700',
};

/**
 * AdminCompanyApplicationsPage — `/admin/applications`。super_admin 専用。
 * 公開フォームから届いた企業の利用申請を一覧し、承認 / 却下する。
 * 承認しても会社は自動作成されないため、承認後は「招待管理」から会社管理者を招待する。
 */
export default function AdminCompanyApplicationsPage() {
  const isAdmin = useAppSelector((state) => state.auth.isAdmin);
  const authLoading = useAppSelector((state) => state.auth.loading);
  const role = useAppSelector((state) => state.auth.role);
  const { applications, pendingCount, loading, error, updatingId, setStatus } =
    useCompanyApplications();
  const { showToast } = useToast();

  if (authLoading) return <Loading message="認証情報を確認中..." className="min-h-[50vh]" />;
  // 利用申請は全テナント横断の運営機能なので super_admin のみ。
  if (!isAdmin || role !== 'super_admin') return <Navigate to="/" replace />;

  const update = async (app: CompanyApplication, status: CompanyApplicationStatus) => {
    const ok = await setStatus(app.id, status);
    if (ok) {
      showToast(
        'success',
        status === 'approved'
          ? `「${app.companyName}」を承認しました。招待管理から会社管理者を招待してください。`
          : `「${app.companyName}」を却下しました。`,
      );
    }
  };

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="管理: 利用申請"
        description="公開フォームから届いた企業の利用申請です。内容を確認して承認または却下します。承認しても会社は自動作成されないため、承認後は「招待管理」から会社管理者を招待してください。"
      />

      {error && (
        <div role="alert" className="p-3 rounded border border-rose-300 bg-rose-50 text-rose-800 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <Loading message="読み込み中..." className="min-h-[30vh]" />
      ) : applications.length === 0 ? (
        <p className="text-sm text-[var(--color-text-muted)]">利用申請はまだありません。</p>
      ) : (
        <>
          <p className="text-sm text-[var(--color-text-secondary)]">
            承認待ち:{' '}
            <span className="font-semibold text-amber-700">{pendingCount}</span> 件 / 全{' '}
            {applications.length} 件
          </p>

          <ul className="space-y-3">
            {applications.map((app) => (
              <li
                key={app.id}
                className="p-4 border rounded-lg bg-[var(--color-surface-1)] space-y-2"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex items-center gap-3 min-w-0">
                    <BuildingOffice2Icon className="w-7 h-7 text-[var(--color-text-muted)] flex-shrink-0" />
                    <div className="min-w-0">
                      <p className="font-semibold text-sm flex items-center gap-2 flex-wrap">
                        <span className="truncate">{app.companyName}</span>
                        <span
                          className={`text-[10px] font-medium px-1.5 py-0.5 rounded ${STATUS_CLASS[app.status]}`}
                        >
                          {STATUS_LABEL[app.status]}
                        </span>
                      </p>
                      <p className="text-xs text-[var(--color-text-muted)] mt-0.5 truncate">
                        {app.applicantName}（{app.email}）
                      </p>
                      <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                        申請日: {new Date(app.createdAt).toLocaleDateString('ja-JP')}
                      </p>
                    </div>
                  </div>

                  {app.status === 'pending' && (
                    <div className="flex gap-2 flex-shrink-0">
                      <button
                        type="button"
                        onClick={() => update(app, 'approved')}
                        disabled={updatingId === app.id}
                        className="flex items-center gap-1 text-xs px-3 py-1.5 rounded border border-emerald-300 text-emerald-700 hover:bg-emerald-50 transition-colors disabled:opacity-50"
                      >
                        <CheckIcon className="w-3.5 h-3.5" />
                        承認
                      </button>
                      <button
                        type="button"
                        onClick={() => {
                          if (window.confirm(`「${app.companyName}」の申請を却下しますか？`)) {
                            update(app, 'rejected');
                          }
                        }}
                        disabled={updatingId === app.id}
                        className="flex items-center gap-1 text-xs px-3 py-1.5 rounded border border-rose-300 text-rose-700 hover:bg-rose-50 transition-colors disabled:opacity-50"
                      >
                        <XMarkIcon className="w-3.5 h-3.5" />
                        却下
                      </button>
                    </div>
                  )}

                  {app.status === 'approved' && (
                    <Link
                      to="/admin/invitations"
                      className="flex items-center gap-1 text-xs px-3 py-1.5 border rounded text-[var(--color-text-secondary)] hover:bg-surface-2 transition-colors flex-shrink-0"
                    >
                      <UserPlusIcon className="w-3.5 h-3.5" />
                      招待へ
                    </Link>
                  )}
                </div>

                {app.message && (
                  <p className="text-xs text-[var(--color-text-secondary)] whitespace-pre-wrap border-t border-surface-3 pt-2">
                    {app.message}
                  </p>
                )}
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
}
