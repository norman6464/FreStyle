import { useEffect, Suspense } from 'react';
import { Routes, Route, useLocation } from 'react-router-dom';
import AuthInitializer from './providers/AuthInitializer';
import Protected from './providers/Protected';
import { AppShell } from '@/widgets/app-shell';
import ErrorBoundary from './providers/ErrorBoundary';
import Loading from '@/components/Loading';
import { MaintenancePage } from '@/pages/maintenance';
import { ToastProvider } from './providers/ToastProvider';
import { useToast } from '@/hooks/useToast';
import { useBackendHealth } from '@/hooks/useBackendHealth';
import ToastContainer from '@/components/ToastContainer';
import { lazyWithReload, clearLazyReloadFlags } from '@/shared/lib/lazyWithReload';

// 認証不要ページ
const LoginPage = lazyWithReload(() => import('@/pages/login').then((m) => ({ default: m.LoginPage })), 'LoginPage');
const LoginCallback = lazyWithReload(() => import('@/pages/login-callback').then((m) => ({ default: m.LoginCallback })), 'LoginCallback');
const ForgotPasswordPage = lazyWithReload(() => import('@/pages/forgot-password').then((m) => ({ default: m.ForgotPasswordPage })), 'ForgotPasswordPage');
const ConfirmForgotPasswordPage = lazyWithReload(() => import('@/pages/confirm-forgot-password').then((m) => ({ default: m.ConfirmForgotPasswordPage })), 'ConfirmForgotPasswordPage');
const AcceptInvitationPage = lazyWithReload(() => import('@/pages/accept-invitation').then((m) => ({ default: m.AcceptInvitationPage })), 'AcceptInvitationPage');
const CompanyApplicationPage = lazyWithReload(() => import('@/pages/company-application').then((m) => ({ default: m.CompanyApplicationPage })), 'CompanyApplicationPage');

// 認証必要ページ
const MenuPage = lazyWithReload(() => import('@/pages/home').then((m) => ({ default: m.MenuPage })), 'MenuPage');
const SettingsPage = lazyWithReload(() => import('@/pages/settings').then((m) => ({ default: m.SettingsPage })), 'SettingsPage');
const AskAiPage = lazyWithReload(() => import('@/pages/ask-ai').then((m) => ({ default: m.AskAiPage })), 'AskAiPage');
const NotesPage = lazyWithReload(() => import('@/pages/notes').then((m) => ({ default: m.NotesPage })), 'NotesPage');
const NotificationPage = lazyWithReload(() => import('@/pages/notifications').then((m) => ({ default: m.NotificationPage })), 'NotificationPage');
const LearningReportPage = lazyWithReload(() => import('@/pages/learning-report').then((m) => ({ default: m.LearningReportPage })), 'LearningReportPage');
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
const CoursesListPage = lazyWithReload(() => import('@/pages/courses').then((m) => ({ default: m.CoursesListPage })), 'CoursesListPage');
const CourseDetailPage = lazyWithReload(() => import('@/pages/course-detail').then((m) => ({ default: m.CourseDetailPage })), 'CourseDetailPage');
const MarkdownSyntaxHelpPage = lazyWithReload(() => import('@/pages/markdown-syntax-help').then((m) => ({ default: m.MarkdownSyntaxHelpPage })), 'MarkdownSyntaxHelpPage');
// inkwell プリミティブの見た目確認用カタログ（認証不要・削除可）。
const InkwellShowcasePage = lazyWithReload(() => import('@/pages/inkwell-showcase').then((m) => ({ default: m.InkwellShowcasePage })), 'InkwellShowcasePage');

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

export default function App() {
  const { status: healthStatus } = useBackendHealth();

  // バックエンドが連続失敗で unhealthy になっているとメンテナンスページを表示。
  // healthy / unknown は通常通りアプリを描画（unknown は初回 health check 完了前で、
  // ここで loading を出すと体感が遅くなるので楽観的にアプリを表示する）。
  if (healthStatus === 'unhealthy') {
    return (
      <ErrorBoundary>
        <MaintenancePage />
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
    <ToastProvider>
    <Suspense fallback={<Loading fullscreen message="読み込み中..." />}>
    <Routes>
      {/* 誰でもアクセス可能 */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/login/callback" element={<LoginCallback />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route
        path="/confirm-forgot-password"
        element={<ConfirmForgotPasswordPage />}
      />
      {/* 招待マジックリンクの受諾画面（認証不要・SES メールから踏まれる） */}
      <Route path="/invitations/accept" element={<AcceptInvitationPage />} />
      <Route path="/company-application" element={<CompanyApplicationPage />} />
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
        <Route path="/" element={<MenuPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        {/* 旧 /profile/me は /settings に統合（後方互換のため redirect 相当として SettingsPage を出す） */}
        <Route path="/profile/me" element={<SettingsPage />} />
        <Route path="/chat/ask-ai" element={<AskAiPage />} />
        <Route path="/chat/ask-ai/:sessionId" element={<AskAiPage />} />
        <Route path="/notes" element={<NotesPage />} />
        <Route path="/notes/markdown-help" element={<MarkdownSyntaxHelpPage />} />
        <Route path="/notifications" element={<NotificationPage />} />
        <Route path="/reports" element={<LearningReportPage />} />
        <Route path="/help" element={<HelpPage />} />
        {/* コード学習は「言語選択カード → その言語の問題一覧 → 問題」の 3 段(FRESTYLE-152)。
            /lang/:language は 2 セグメントなので 1 セグメントの :slug とは衝突しない。 */}
        <Route path="/code-editor" element={<ExerciseLanguageSelectPage />} />
        <Route path="/code-editor/lang/:language" element={<ExerciseListPage />} />
        <Route path="/code-editor/:slug" element={<ExerciseDetailPage />} />
        <Route path="/courses" element={<CoursesListPage />} />
        <Route path="/courses/:id" element={<CourseDetailPage />} />
        {/* 旧 /teaching-materials へのアクセスは /courses に redirect */}
        <Route path="/teaching-materials" element={<CoursesListPage />} />
        {/* Admin 専用（コンポーネント側で isAdmin チェック → 非 admin は / にリダイレクト） */}
        <Route path="/admin/dashboard" element={<AdminDashboardPage />} />
        <Route path="/admin/companies" element={<AdminCompaniesPage />} />
        <Route path="/admin/applications" element={<AdminCompanyApplicationsPage />} />
        <Route path="/admin/members" element={<AdminMembersPage />} />
        <Route path="/admin/audit" element={<AdminAuditLogPage />} />
        <Route path="/admin/invitations" element={<AdminInvitationsPage />} />
      </Route>
    </Routes>
    </Suspense>
    <NavigationToast />
    <ToastContainer />
    </ToastProvider>
    </ErrorBoundary>
  );
}

