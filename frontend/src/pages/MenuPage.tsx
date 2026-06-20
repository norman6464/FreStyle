import { Link } from 'react-router-dom';
import { useSelector } from 'react-redux';
import {
  ChatBubbleBottomCenterTextIcon,
  CodeBracketIcon,
  DocumentTextIcon,
  DocumentChartBarIcon,
  BuildingOffice2Icon,
  EnvelopeIcon,
  BookOpenIcon,
  ArrowRightIcon,
} from '@heroicons/react/24/outline';
import type { RootState } from '../store';
import { useUserDashboard } from '../hooks/useUserDashboard';
import DashboardStats from '../components/dashboard/DashboardStats';

/**
 * ホーム画面（ダッシュボード）。
 *
 * ロール別にカードセットを出し分け:
 *   - super_admin   : 管理系のみ
 *   - company_admin : 管理 + 学習機能（AI はテナント設定に関わらず常時表示）
 *   - trainee       : 学習機能のみ。aiChatEnabledForTrainees が false なら AI カードを非表示
 *
 * API 依存なし（auth store のみ参照）。 将来的に「最近の演習」等の Quick Stats を
 * 追加する場合はここに useXxxStats() を差し込む。
 */
export default function MenuPage() {
  const role = useSelector((state: RootState) => state.auth.role);
  const aiEnabled = useSelector((state: RootState) => state.auth.aiChatEnabledForTrainees);
  const isSuperAdmin = role === 'super_admin';
  const isTrainee = role === 'trainee';
  const showAi = !isTrainee || aiEnabled;

  // ダッシュボード統計（super_admin は学習機能を使わないので取得しない）。
  const { dashboard } = useUserDashboard({ enabled: !isSuperAdmin });

  return (
    <div className="px-4 sm:px-6 pt-8 pb-24 max-w-6xl mx-auto">
      <div className="flex flex-col lg:flex-row gap-8 items-start">

        {/* ── 左メインコンテンツ ── */}
        <div className="flex-1 min-w-0 space-y-10">

          {/* ウェルカムセクション */}
          <section>
            <p className="text-xs font-semibold text-brand-500 uppercase tracking-widest mb-1">
              {isSuperAdmin ? '運営管理者ダッシュボード' : isTrainee ? '学習ダッシュボード' : '管理者ダッシュボード'}
            </p>
            <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">
              {isSuperAdmin ? '管理メニュー' : 'FreStyle へようこそ'}
            </h1>
            <p className="mt-2 text-sm text-[var(--color-text-muted)]">
              {isSuperAdmin
                ? '企業管理・招待などの運営操作を行えます。'
                : 'コースや演習で学習を進め、AI チャットで疑問を解決しましょう。'}
            </p>
          </section>

          {isSuperAdmin ? (
            <FeatureSection title="管理機能">
              <FeatureCard
                to="/admin/companies"
                icon={BuildingOffice2Icon}
                title="会社一覧"
                description="登録済み企業の管理・閲覧を行います。"
                color="blue"
              />
              <FeatureCard
                to="/admin/invitations"
                icon={EnvelopeIcon}
                title="招待管理"
                description="企業管理者への招待を作成・管理できます。"
                color="blue"
              />
            </FeatureSection>
          ) : (
            <>
              <FeatureSection title="学習">
                <FeatureCard
                  to="/courses"
                  icon={BookOpenIcon}
                  title="コース"
                  description="体系的なカリキュラムで段階的に学べます。"
                  color="emerald"
                  badge="おすすめ"
                />
                <FeatureCard
                  to="/code-editor"
                  icon={CodeBracketIcon}
                  title="コード演習"
                  description="実際にコードを書いて手を動かしながら学べます。"
                  color="emerald"
                />
              </FeatureSection>

              <FeatureSection title="ツール">
                {showAi && (
                  <FeatureCard
                    to="/chat/ask-ai"
                    icon={ChatBubbleBottomCenterTextIcon}
                    title="AI チャット"
                    description="質問・要約・コード補助など自由に対話できます。"
                    color="brand"
                  />
                )}
                <FeatureCard
                  to="/notes"
                  icon={DocumentTextIcon}
                  title="ノート"
                  description="学習メモを書き留め、いつでも振り返れます。"
                  color="taupe"
                />
                <FeatureCard
                  to="/reports"
                  icon={DocumentChartBarIcon}
                  title="学習レポート"
                  description="月次の学習サマリーを確認できます。"
                  color="taupe"
                />
              </FeatureSection>

              {role === 'company_admin' && (
                <FeatureSection title="管理">
                  <FeatureCard
                    to="/admin/members"
                    icon={BuildingOffice2Icon}
                    title="従業員一覧"
                    description="所属メンバーの学習状況を確認できます。"
                    color="blue"
                  />
                  <FeatureCard
                    to="/admin/invitations"
                    icon={EnvelopeIcon}
                    title="招待管理"
                    description="メンバーへの招待を作成・管理できます。"
                    color="blue"
                  />
                </FeatureSection>
              )}
            </>
          )}
        </div>

        {/* ── 右サイドバー（学習統計）── super_admin には非表示 */}
        {!isSuperAdmin && dashboard && (
          <div className="w-full lg:w-72 shrink-0">
            <DashboardStats dashboard={dashboard} />
          </div>
        )}

      </div>
    </div>
  );
}

// ─── サブコンポーネント ──────────────────────────────────────────────────────

interface FeatureSectionProps {
  title: string;
  children: React.ReactNode;
}

function FeatureSection({ title, children }: FeatureSectionProps) {
  return (
    <section className="space-y-3">
      <h2 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-widest">
        {title}
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {children}
      </div>
    </section>
  );
}

type CardColor = 'brand' | 'emerald' | 'taupe' | 'blue';

interface FeatureCardProps {
  to: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  title: string;
  description: string;
  color: CardColor;
  badge?: string;
}

const iconBg: Record<CardColor, string> = {
  brand:   'bg-brand-100 text-brand-600',
  emerald: 'bg-emerald-100 text-emerald-700',
  taupe:   'bg-taupe-100 text-taupe-600',
  blue:    'bg-blue-100 text-blue-600',
};

function FeatureCard({ to, icon: Icon, title, description, color, badge }: FeatureCardProps) {
  return (
    <Link
      to={to}
      className="group relative flex flex-col gap-4 p-5 rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] hover:bg-[var(--color-surface-2)] hover:border-[var(--color-text-muted)]/40 hover:shadow-sm transition-all duration-150"
    >
      {badge && (
        <span className="absolute top-3 right-3 text-[10px] font-semibold px-2 pt-px pb-[3px] rounded-full bg-emerald-100 text-emerald-700">
          {badge}
        </span>
      )}
      <div className={`w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 ${iconBg[color]}`}>
        <Icon className="w-5 h-5 -translate-y-px" />
      </div>
      <div className="flex-1 space-y-1">
        <h3 className="font-semibold text-[var(--color-text-primary)] text-sm group-hover:text-brand-500 transition-colors">
          {title}
        </h3>
        <p className="text-xs text-[var(--color-text-muted)] leading-relaxed">
          {description}
        </p>
      </div>
      <div className="flex items-center gap-1 text-xs text-[var(--color-text-muted)] group-hover:text-brand-500 transition-colors">
        <span>開く</span>
        <ArrowRightIcon className="w-3 h-3 translate-x-0 group-hover:translate-x-0.5 transition-transform" />
      </div>
    </Link>
  );
}
