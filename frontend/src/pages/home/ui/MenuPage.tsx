import { useAppSelector } from '@/shared/lib/store';

import {
  ChatBubbleBottomCenterTextIcon,
  CodeBracketIcon,
  DocumentTextIcon,
  DocumentChartBarIcon,
  BuildingOffice2Icon,
  EnvelopeIcon,
  BookOpenIcon,
} from '@heroicons/react/24/outline';

import { useUserDashboard } from '../model/useUserDashboard';
import { useCompanyLearningSummary } from '../model/useCompanyLearningSummary';
import DashboardStats from './DashboardStats';
import CompanyLearningPanel from './CompanyLearningPanel';
import FeatureSection from './FeatureSection';
import FeatureCard from './FeatureCard';
import MenuSkeleton from './MenuSkeleton';
import StatsSkeleton from './StatsSkeleton';

/**
 * ホーム画面（ダッシュボード）。
 *
 * ロール別にカードセットを出し分け:
 *   - super_admin   : 管理系のみ
 *   - company_admin : 管理 + 学習機能（AI はテナント設定に関わらず常時表示）
 *   - trainee       : 学習機能のみ。aiChatEnabledForTrainees が false なら AI カードを非表示
 *
 * 表示タイミング:
 *   学習者向けはメニューカードとパーソナライズ統計（右サイドバー）を **同時に** 出す。
 *   統計 API のロード中はメニューを先出しせずスケルトンで待ち、レイアウトシフトと
 *   「メニューだけ先に出てサイドバーが後から差し込まれる」ちらつきを防ぐ。
 *   super_admin は統計を持たないので即時表示。
 *
 *   role が null の間はどのロールとしても描画しない（FRESTYLE-233）。null は「未認証」と
 *   「未確定」の両方を表すため、確定前に描画すると全ての判定が false になり、既定として
 *   学習者向けが出てしまう。管理者のログイン直後に学習者画面が一瞬映る原因だった。
 */
export default function MenuPage() {
  const role = useAppSelector((state) => state.auth.role);
  const aiEnabled = useAppSelector((state) => state.auth.aiChatEnabledForTrainees);
  const isSuperAdmin = role === 'super_admin';
  const isTrainee = role === 'trainee';
  const isCompanyAdmin = role === 'company_admin';
  const roleUnresolved = role === null;
  const showAi = !isTrainee || aiEnabled;

  // サイドバーの中身はロールで出し分ける(FRESTYLE-103):
  //   trainee = 自分の学習統計 / company_admin = 自社メンバーの学習状況 / super_admin = なし。
  // company_admin は学習者ではないため自分用の /me/dashboard は取得しない。
  const { dashboard, loading: dashboardLoading } = useUserDashboard({ enabled: isTrainee });
  const { summary, loading: summaryLoading } = useCompanyLearningSummary({ enabled: isCompanyAdmin });

  // 学習者/管理者向けはサイドバーのロード完了まで本体を出さず、両カラムを同時に出す。
  const waitingForStats = (isTrainee && dashboardLoading) || (isCompanyAdmin && summaryLoading);

  // ロール未確定のうちは見出しもサイドバーもロールに依存するため、ページ全体を
  // 読み込み表示にする。ここで役割別の要素を出すと、確定後に差し替わってちらつく。
  if (roleUnresolved) {
    return (
      <div className="px-4 sm:px-6 pt-8 pb-24 max-w-6xl mx-auto" aria-busy="true">
        {/* スケルトンは装飾（aria-hidden）なので、待機中であることは live region で伝える。 */}
        <span role="status" className="sr-only">
          読み込み中
        </span>
        <MenuSkeleton />
      </div>
    );
  }

  return (
    <div className="px-4 sm:px-6 pt-8 pb-24 max-w-6xl mx-auto">
      {/* ウェルカムセクション（データ非依存・即時表示） */}
      <section className="mb-8">
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

      <div className="flex flex-col lg:flex-row gap-8 items-start">
        {/* ── 左メインコンテンツ ── */}
        <div className="flex-1 min-w-0 space-y-8 w-full">
          {waitingForStats ? (
            <MenuSkeleton />
          ) : isSuperAdmin ? (
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
                  techLogos={['git', 'go', 'docker', 'php']}
                />
                <FeatureCard
                  to="/code-editor"
                  icon={CodeBracketIcon}
                  title="コード演習"
                  description="実際にコードを書いて手を動かしながら学べます。"
                  color="emerald"
                  techLogos={['go', 'php', 'javascript', 'typescript']}
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

        {/* ── 右サイドバー ── trainee = 自分の学習統計 / company_admin = メンバーの学習状況 /
            super_admin = 非表示 (FRESTYLE-103) */}
        {!isSuperAdmin && (
          <div className="w-full lg:w-72 shrink-0">
            {waitingForStats ? (
              <StatsSkeleton />
            ) : isTrainee ? (
              dashboard && <DashboardStats dashboard={dashboard} />
            ) : (
              summary && <CompanyLearningPanel summary={summary} />
            )}
          </div>
        )}
      </div>
    </div>
  );
}
