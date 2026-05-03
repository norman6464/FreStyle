import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Navigate, Link } from 'react-router-dom';
import CompanyRepository, { Company } from '../repositories/CompanyRepository';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import PageIntro from '../components/ui/PageIntro';
import { logger } from '../lib/logger';
import { BuildingOffice2Icon, UserPlusIcon, Cog6ToothIcon } from '@heroicons/react/24/outline';

export default function AdminCompaniesPage() {
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);
  const authLoading = useSelector((state: RootState) => state.auth.loading);

  const [companies, setCompanies] = useState<Company[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isAdmin) return;
    CompanyRepository.list()
      .then(setCompanies)
      .catch((e) => {
        setError('会社一覧の取得に失敗しました');
        logger.error(e);
      })
      .finally(() => setLoading(false));
  }, [isAdmin]);

  if (authLoading) return <Loading message="認証情報を確認中..." />;
  if (!isAdmin) return <Navigate to="/" replace />;

  return (
    <div className="px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="管理: 会社一覧"
        description="登録されている会社の一覧です。会社を選択してシナリオや招待を管理できます。"
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
                  <p className="font-semibold text-sm">{company.name}</p>
                  <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                    登録日: {new Date(company.createdAt).toLocaleDateString('ja-JP')}
                  </p>
                </div>
              </div>

              <div className="flex gap-2 flex-shrink-0">
                <Link
                  to={`/admin/scenarios?companyId=${company.id}`}
                  className="flex items-center gap-1 text-xs px-3 py-1.5 border rounded text-[var(--color-text-secondary)] hover:bg-surface-2 transition-colors"
                >
                  <Cog6ToothIcon className="w-3.5 h-3.5" />
                  シナリオ
                </Link>
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
