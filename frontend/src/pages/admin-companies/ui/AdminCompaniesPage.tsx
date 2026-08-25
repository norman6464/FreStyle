import { useEffect, useState } from 'react';

import { Link } from 'react-router-dom';
import { CompanyRepository, CompanyStat } from '@/entities/company';

import Loading from '@/shared/ui/Loading';
import FormMessage from '@/shared/ui/FormMessage';
import PageIntro from '@/shared/ui/PageIntro';
import { logger } from '@/shared/lib/logger';
import { BuildingOffice2Icon, UserPlusIcon } from '@heroicons/react/24/outline';

// 会社一覧 / 横断ビュー（/admin/companies/stats）は super_admin 専用エンドポイント。
// ルート側の RequireRole が super_admin 以外を通さないので、ここでは判定しない。
export default function AdminCompaniesPage() {
  const [companies, setCompanies] = useState<CompanyStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  useEffect(() => {
    CompanyRepository.listStats()
      .then(setCompanies)
      .catch((e) => {
        setError('会社一覧の取得に失敗しました');
        logger.error(e);
      })
      .finally(() => setLoading(false));
  }, []);

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

  return (
    <div className="px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="管理: 会社一覧"
        description="登録されている会社の一覧です。各社のアカウントの有効/無効を切り替えたり、招待を管理できます。"
      />

      <FormMessage message={error ? { type: 'error', text: error } : null} />

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
