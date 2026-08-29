import { useEffect, Suspense } from 'react';
import { Routes, Route, useLocation, Navigate, useParams } from 'react-router-dom';
import AuthInitializer from './providers/AuthInitializer';
import Protected from './providers/Protected';
import RequireRole from './providers/RequireRole';
import { AppShell } from '@/widgets/app-shell';
import ErrorBoundary from './providers/ErrorBoundary';
import Loading from '@/shared/ui/Loading';
import { ToastProvider } from './providers/ToastProvider';
import { useToast } from '@/shared/lib/hooks/useToast';
import ToastContainer from '@/app/providers/ToastContainer';
import { lazyWithReload, clearLazyReloadFlags } from '@/shared/lib/lazyWithReload';

/* v8 ignore start -- 以下はコード分割のためのルート表。各 `() => import(...)` は
   中身を持たない読み込み用の関数で、埋めるには全ページを描画するしかなく指標として
   意味を持たない。実ロジック（NavigationToast / AppRoutes）は計測対象のまま残す。 */
// 認証不要ページ
const LoginPage = lazyWithReload(() => import('@/pages/login').then((m) => ({ default: m.LoginPage })), 'LoginPage');
const SignupPage = lazyWithReload(() => import('@/pages/signup').then((m) => ({ default: m.SignupPage })), 'SignupPage');
const LoginCallback = lazyWithReload(() => import('@/pages/login-callback').then((m) => ({ default: m.LoginCallback })), 'LoginCallback');
const ForgotPasswordPage = lazyWithReload(() => import('@/pages/forgot-password').then((m) => ({ default: m.ForgotPasswordPage })), 'ForgotPasswordPage');
const ConfirmForgotPasswordPage = lazyWithReload(() => import('@/pages/confirm-forgot-password').then((m) => ({ default: m.ConfirmForgotPasswordPage })), 'ConfirmForgotPasswordPage');
const AcceptInvitationPage = lazyWithReload(() => import('@/pages/accept-invitation').then((m) => ({ default: m.AcceptInvitationPage })), 'AcceptInvitationPage');
// 公開ランディング（SEO 対象・認証不要）。ログイン済みは /dashboard へ送る。
const LandingPage = lazyWithReload(() => import('@/pages/landing').then((m) => ({ default: m.LandingPage })), 'LandingPage');

// 認証必要ページ
const MenuPage = lazyWithReload(() => import('@/pages/home').then((m) => ({ default: m.MenuPage })), 'MenuPage');
const SettingsPage = lazyWithReload(() => import('@/pages/settings').then((m) => ({ default: m.SettingsPage })), 'SettingsPage');
const NotePage = lazyWithReload(() => import('@/pages/note').then((m) => ({ default: m.NotePage })), 'NotePage');
const NotificationPage = lazyWithReload(() => import('@/pages/notifications').then((m) => ({ default: m.NotificationPage })), 'NotificationPage');
const HelpPage = lazyWithReload(() => import('@/pages/help').then((m) => ({ default: m.HelpPage })), 'HelpPage');
const AdminInvitationsPage = lazyWithReload(() => import('@/pages/admin-invitations').then((m) => ({ default: m.AdminInvitationsPage })), 'AdminInvitationsPage');
const AdminCompaniesPage = lazyWithReload(() => import('@/pages/admin-companies').then((m) => ({ default: m.AdminCompaniesPage })), 'AdminCompaniesPage');
const AdminMembersPage = lazyWithReload(() => import('@/pages/admin-members').then((m) => ({ default: m.AdminMembersPage })), 'AdminMembersPage');
const AdminCompanyApplicationsPage = lazyWithReload(
  () => import('@/pages/admin-company-applications').then((m) => ({ default: m.AdminCompanyApplicationsPage })),
  'AdminCompanyApplicationsPage',
);
const AdminDashboardPage = lazyWithReload(() => import('@/pages/admin-dashboard').then((m) => ({ default: m.AdminDashboardPage })), 'AdminDashboardPage');
const AdminAuditLogPage = lazyWithReload(() => import('@/pages/admin-audit-log').then((m) => ({ default: m.AdminAuditLogPage })), 'AdminAuditLogPage');
const ExerciseLanguageSelectPage = lazyWithReload(() => import('@/pages/exercise-languages').then((m) => ({ default: m.ExerciseLanguageSelectPage })), 'ExerciseLanguageSelectPage');
const ExerciseListPage = lazyWithReload(() => import('@/pages/exercises').then((m) => ({ default: m.ExerciseListPage })), 'ExerciseListPage');
const ExerciseDetailPage = lazyWithReload(() => import('@/pages/exercise-detail').then((m) => ({ default: m.ExerciseDetailPage })), 'ExerciseDetailPage');
const CourseCategorySelectPage = lazyWithReload(() => import('@/pages/courses').then((m) => ({ default: m.CourseCategorySelectPage })), 'CourseCategorySelectPage');
const CoursesListPage = lazyWithReload(() => import('@/pages/courses').then((m) => ({ default: m.CoursesListPage })), 'CoursesListPage');
const CourseDetailPage = lazyWithReload(() => import('@/pages/course-detail').then((m) => ({ default: m.CourseDetailPage })), 'CourseDetailPage');
// inkwell プリミティブの見た目確認用カタログ（認証不要・削除可）。
const InkwellShowcasePage = lazyWithReload(() => import('@/pages/inkwell-showcase').then((m) => ({ default: m.InkwellShowcasePage })), 'InkwellShowcasePage');
const NotFoundPage = lazyWithReload(() => import('@/pages/not-found').then((m) => ({ default: m.NotFoundPage })), 'NotFoundPage');
/* v8 ignore stop */

function NavigationToast() {
  const location = useLocation();
  const { showToast } = useToast();

  useEffect(() => {
    const toast = (location.state as { toast?: string })?.toast;
    if (toast) {
      showToast('success', toast);
      window.history.replaceState({}, '');
    }
    // ナビゲーション成功 = 直前の lazy reload で復旧した。次回また chunk が
    // 失敗したら再度 reload を許可するため、フラグをクリアしておく。
    clearLazyReloadFlags();
  }, [location, showToast]);

  return null;
}

// LegacyKbPageRedirect は旧 /kb/:slug/pages/:pageId を /p/:pageId へ写す。
// slug は URL から消えた（テナントはページ ID から解決する）ので捨ててよい。
function LegacyKbPageRedirect() {
  const { pageId } = useParams<{ pageId: string }>();
  return <Navigate to={`/p/${pageId ?? ''}`} replace />;
}

export default function App() {
  return (
    <ErrorBoundary>
    <ToastProvider>
    <Suspense fallback={<Loading fullscreen message="読み込み中..." />}>
    <Routes>
      {/* 誰でもアクセス可能 */}
      {/* 公開トップ（SEO 対象）。ログイン済みは LandingPage 内で /dashboard へ送る。 */}
      <Route path="/" element={<LandingPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/signup" element={<SignupPage />} />
      <Route path="/login/callback" element={<LoginCallback />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route
        path="/confirm-forgot-password"
        element={<ConfirmForgotPasswordPage />}
      />
      {/* 招待マジックリンクの受諾画面（認証不要・SES メールから踏まれる） */}
      <Route path="/invitations/accept" element={<AcceptInvitationPage />} />
      {/* inkwell UI カタログ（見た目確認用・認証不要） */}
      <Route path="/dev/inkwell" element={<InkwellShowcasePage />} />

      {/* 認証が必要（AppShell レイアウト内） */}
      <Route
        element={
          <AuthInitializer>
            <Protected>
              <AppShell />
            </Protected>
          </AuthInitializer>
        }
      >
        {/* ログイン後のダッシュボード。旧 "/" から移設（"/" は公開 LP に）。 */}
        <Route path="/dashboard" element={<MenuPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        {/* 旧 /profile/me は /settings に統合（後方互換のため redirect 相当として SettingsPage を出す） */}
        <Route path="/profile/me" element={<SettingsPage />} />
        {/*
          ノート（workspaces → spaces → pages の木）。旧「ナレッジ」を統合した現在の正
          （FRESTYLE-393）。ページの URL は /p/{pageId} だけで、テナントを URL に出さない。
        */}
        <Route path="/notes" element={<NotePage />} />
        <Route path="/p/:pageId" element={<NotePage />} />
        {/* 旧 URL の受け皿。/kb 系はブックマーク・共有リンクから来る。 */}
        <Route path="/kb" element={<Navigate to="/notes" replace />} />
        <Route path="/kb/:workspaceSlug" element={<Navigate to="/notes" replace />} />
        <Route path="/kb/:workspaceSlug/pages/:pageId" element={<LegacyKbPageRedirect />} />
        <Route path="/notifications" element={<NotificationPage />} />
        <Route path="/help" element={<HelpPage />} />
        {/* コード学習は「言語選択カード → その言語の問題一覧 → 問題」の 3 段(FRESTYLE-152)。
            /lang/:language は 2 セグメントなので 1 セグメントの :slug とは衝突しない。 */}
        <Route path="/code-editor" element={<ExerciseLanguageSelectPage />} />
        <Route path="/code-editor/lang/:language" element={<ExerciseListPage />} />
        <Route path="/code-editor/:slug" element={<ExerciseDetailPage />} />
        {/* コースは「学習領域の選択 → その領域の一覧 → 詳細」の 3 段(FRESTYLE-177)。
            /category/:category は 2 セグメントなので 1 セグメントの :id とは衝突しない。 */}
        <Route path="/courses" element={<CourseCategorySelectPage />} />
        <Route path="/courses/category/:category" element={<CoursesListPage />} />
        <Route path="/courses/:id" element={<CourseDetailPage />} />
        {/* 旧 /teaching-materials へのアクセスは /courses に redirect */}
        <Route path="/teaching-materials" element={<CourseCategorySelectPage />} />
        {/* Admin 専用。通過条件はここ（RequireRole）に集約する → 満たさなければ /dashboard へ。
            画面ごとに条件が違うのは現行のまま: 会社一覧 / 監査ログは role のみ、
            運営ダッシュボード / 利用申請は role + isAdmin、従業員一覧 / 招待は isAdmin のみ。 */}
        <Route
          path="/admin/dashboard"
          element={
            <RequireRole allow={['super_admin']} requireAdminFlag>
              <AdminDashboardPage />
            </RequireRole>
          }
        />
        <Route
          path="/admin/companies"
          element={
            <RequireRole allow={['super_admin']}>
              <AdminCompaniesPage />
            </RequireRole>
          }
        />
        <Route
          path="/admin/applications"
          element={
            <RequireRole allow={['super_admin']} requireAdminFlag>
              <AdminCompanyApplicationsPage />
            </RequireRole>
          }
        />
        <Route
          path="/admin/members"
          element={
            <RequireRole allow="any" requireAdminFlag>
              <AdminMembersPage />
            </RequireRole>
          }
        />
        <Route
          path="/admin/audit"
          element={
            <RequireRole allow={['super_admin']}>
              <AdminAuditLogPage />
            </RequireRole>
          }
        />
        <Route
          path="/admin/invitations"
          element={
            <RequireRole allow="any" requireAdminFlag>
              <AdminInvitationsPage />
            </RequireRole>
          }
        />
      </Route>

      {/* どのルートにも一致しない URL の受け皿（FRESTYLE-86）。
          認証ブロックの外に置く: 中に入れると未ログイン時に /login へ飛ばされ、
          タイポや古いリンクで来た訪問者に 404 を見せられない（公開サイトとして不適切）。 */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
    </Suspense>
    <NavigationToast />
    <ToastContainer />
    </ToastProvider>
    </ErrorBoundary>
  );
}

