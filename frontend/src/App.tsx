import { useEffect, Suspense } from 'react';
import { Routes, Route, useLocation } from 'react-router-dom';
import AuthInitializer from './utils/AuthInitializer';
import Protected from './utils/Protected';
import AppShell from './components/layout/AppShell';
import ErrorBoundary from './components/ErrorBoundary';
import Loading from './components/Loading';
import MaintenancePage from './pages/MaintenancePage';
import { ToastProvider } from './components/ToastProvider';
import { useToast } from './hooks/useToast';
import { useBackendHealth } from './hooks/useBackendHealth';
import ToastContainer from './components/ToastContainer';
import { lazyWithReload, clearLazyReloadFlags } from './utils/lazyWithReload';

// 認証不要ページ
const LoginPage = lazyWithReload(() => import('./pages/LoginPage'), 'LoginPage');
const SignupPage = lazyWithReload(() => import('./pages/SignupPage'), 'SignupPage');
const LoginCallback = lazyWithReload(() => import('./pages/LoginCallback'), 'LoginCallback');
const ConfirmPage = lazyWithReload(() => import('./pages/ConfirmPage'), 'ConfirmPage');
const ForgotPasswordPage = lazyWithReload(() => import('./pages/ForgotPasswordPage'), 'ForgotPasswordPage');
const ConfirmForgotPasswordPage = lazyWithReload(() => import('./pages/ConfirmForgotPasswordPage'), 'ConfirmForgotPasswordPage');
const AcceptInvitationPage = lazyWithReload(() => import('./pages/AcceptInvitationPage'), 'AcceptInvitationPage');
const WelcomePage = lazyWithReload(() => import('./pages/WelcomePage'), 'WelcomePage');

// 認証必要ページ
const MenuPage = lazyWithReload(() => import('./pages/MenuPage'), 'MenuPage');
const ProfilePage = lazyWithReload(() => import('./pages/ProfilePage'), 'ProfilePage');
const AskAiPage = lazyWithReload(() => import('./pages/AskAiPage'), 'AskAiPage');
const NotesPage = lazyWithReload(() => import('./pages/NotesPage'), 'NotesPage');
const NotificationPage = lazyWithReload(() => import('./pages/NotificationPage'), 'NotificationPage');
const LearningReportPage = lazyWithReload(() => import('./pages/LearningReportPage'), 'LearningReportPage');
const HelpPage = lazyWithReload(() => import('./pages/HelpPage'), 'HelpPage');
const AdminInvitationsPage = lazyWithReload(() => import('./pages/AdminInvitationsPage'), 'AdminInvitationsPage');
const AdminCompaniesPage = lazyWithReload(() => import('./pages/AdminCompaniesPage'), 'AdminCompaniesPage');
const ExerciseListPage = lazyWithReload(() => import('./pages/ExerciseListPage'), 'ExerciseListPage');
const ExerciseDetailPage = lazyWithReload(() => import('./pages/ExerciseDetailPage'), 'ExerciseDetailPage');
const TeachingMaterialsPage = lazyWithReload(() => import('./pages/TeachingMaterialsPage'), 'TeachingMaterialsPage');

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
  const { status: healthStatus, recheck } = useBackendHealth();

  // バックエンドが連続失敗で unhealthy になっているとメンテナンスページを表示。
  // healthy / unknown は通常通りアプリを描画（unknown は初回 health check 完了前で、
  // ここで loading を出すと体感が遅くなるので楽観的にアプリを表示する）。
  if (healthStatus === 'unhealthy') {
    return (
      <ErrorBoundary>
        <MaintenancePage onRetry={recheck} />
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
      <Route path="/signup" element={<SignupPage />} />
      <Route path="/confirm" element={<ConfirmPage />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route
        path="/confirm-forgot-password"
        element={<ConfirmForgotPasswordPage />}
      />
      {/* 招待マジックリンクの受諾画面（認証不要・SES メールから踏まれる） */}
      <Route path="/invitations/accept" element={<AcceptInvitationPage />} />

      {/* Welcome / オンボーディング（認証必須・AppShell 外で全画面表示） */}
      <Route
        path="/welcome"
        element={
          <AuthInitializer>
            <Protected>
              <WelcomePage />
            </Protected>
          </AuthInitializer>
        }
      />

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
        <Route path="/profile/me" element={<ProfilePage />} />
        <Route path="/chat/ask-ai" element={<AskAiPage />} />
        <Route path="/chat/ask-ai/:sessionId" element={<AskAiPage />} />
        <Route path="/notes" element={<NotesPage />} />
        <Route path="/notifications" element={<NotificationPage />} />
        <Route path="/reports" element={<LearningReportPage />} />
        <Route path="/help" element={<HelpPage />} />
        <Route path="/code-editor" element={<ExerciseListPage />} />
        <Route path="/code-editor/:slug" element={<ExerciseDetailPage />} />
        <Route path="/teaching-materials" element={<TeachingMaterialsPage />} />
        {/* Admin 専用（コンポーネント側で isAdmin チェック → 非 admin は / にリダイレクト） */}
        <Route path="/admin/companies" element={<AdminCompaniesPage />} />
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

