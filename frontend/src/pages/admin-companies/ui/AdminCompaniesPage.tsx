import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Navigate, Link } from 'react-router-dom';
import CompanyRepository, { CompanyStat } from '@/repositories/CompanyRepository';
import type { RootState } from '@/store';
import Loading from '@/components/Loading';
import PageIntro from '@/shared/ui/PageIntro';
import { logger } from '@/shared/lib/logger';
import { BuildingOffice2Icon, UserPlusIcon } from '@heroicons/react/24/outline';

export default function AdminCompaniesPage() {
  const authLoading = useSelector((state: RootState) => state.auth.loading);
  const role = useSelector((state: RootState) => state.auth.role);
  // 会社一覧 / 横断ビュー（/admin/companies/stats）は super_admin 専用エンドポイント。
  // company_admin が到達しても 403 を踏ませないよう、判定は super_admin に統一する。
  const isSuperAdmin = role === 'super_admin';

  const [companies, setCompanies] = useState<CompanyStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  useEffect(() => {
    if (!isSuperAdmin) return;
    CompanyRepository.listStats()
      .then(setCompanies)
      .catch((e) => {
        setError('会社一覧の取得に失敗しました');
        logger.error(e);
      })
      .finally(() => setLoading(false));
  }, [isSuperAdmin]);

  // 会社アカウントの有効/無効を切り替える（super_admin 専用）。楽観的更新 + 失敗時ロールバック。
  const setActive = async (id: number, active: boolean) => {
    setUpdatingId(id);
    setError(null);
    setCompanies((prev) => prev.map((c) => (c.id === id ? { ...c, isActive: active } : c)));
    try {
      await CompanyRepository.updateActive(id, active);
    } catch (e) {
      logger.error(e);
      setCompanies((prev) => prev.map((c) => (c.id === id ? { ...c, isActive: !active } : c)));
      setError('会社状態の更新に失敗しました');
    } finally {
      setUpdatingId(null);
    }
  };

  if (authLoading) return <Loading message="認証情報を確認中..." />;
  if (!isSuperAdmin) return <Navigate to="/" replace />;

  return (
    <div className="px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="管理: 会社一覧"
        description="登録されている会社の一覧です。各社のアカウントの有効/無効を切り替えたり、招待を管理できます。"
      />

      {error && (
        <div role="alert" className="p-3 rounded border border-red-300 bg-red-50 text-red-800 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <Loading message="読み込み中..." />
      ) : companies.length === 0 ? (
        <p className="text-sm text-[var(--color-text-muted)]">会社が登録されていません</p>
      ) : (
        <ul className="space-y-3">
          {companies.map((company) => (
            <li
              key={company.id}
              className="p-4 border rounded-lg bg-[var(--color-surface-1)] flex items-start justify-between gap-4"
            >
              <div className="flex items-center gap-3 flex-1">
                <BuildingOffice2Icon className="w-8 h-8 text-[var(--color-text-muted)] flex-shrink-0" />
                <div>
                  <p className="font-semibold text-sm flex items-center gap-2">
                    {company.name}
                    {!company.isActive && (
                      <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-rose-100 text-rose-700">
                        無効
                      </span>
                    )}
                  </p>
                  <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                    メンバー {company.memberTotal}（有効 {company.activeMembers} / 受講者{' '}
                    {company.traineeCount}）
                  </p>
                  <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                    登録日: {new Date(company.createdAt).toLocaleDateString('ja-JP')}
                  </p>
                </div>
              </div>

              <div className="flex gap-2 flex-shrink-0 items-center">
                {isSuperAdmin && (
                  <button
                    type="button"
                    onClick={() => setActive(company.id, !company.isActive)}
                    disabled={updatingId === company.id}
                    className={`text-xs px-3 py-1.5 rounded border transition-colors disabled:opacity-50 ${
                      company.isActive
                        ? 'border-rose-300 text-rose-700 hover:bg-rose-50'
                        : 'border-emerald-300 text-emerald-700 hover:bg-emerald-50'
                    }`}
                  >
                    {company.isActive ? '無効化' : '有効化'}
                  </button>
                )}
                <Link
                  to={`/admin/invitations?companyId=${company.id}`}
                  className="flex items-center gap-1 text-xs px-3 py-1.5 border rounded text-[var(--color-text-secondary)] hover:bg-surface-2 transition-colors"
                >
                  <UserPlusIcon className="w-3.5 h-3.5" />
                  招待
                </Link>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
