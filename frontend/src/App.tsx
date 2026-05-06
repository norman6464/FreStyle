import { useEffect, lazy, Suspense } from 'react';
import { Routes, Route, useLocation } from 'react-router-dom';
import AuthInitializer from './utils/AuthInitializer';
import Protected from './utils/Protected';
import AppShell from './components/layout/AppShell';
import ErrorBoundary from './components/ErrorBoundary';
import Loading from './components/Loading';
import { ToastProvider } from './components/ToastProvider';
import { useToast } from './hooks/useToast';
import ToastContainer from './components/ToastContainer';

// 認証不要ページ
const LoginPage = lazy(() => import('./pages/LoginPage'));
const SignupPage = lazy(() => import('./pages/SignupPage'));
const LoginCallback = lazy(() => import('./pages/LoginCallback'));
const ConfirmPage = lazy(() => import('./pages/ConfirmPage'));
const ForgotPasswordPage = lazy(() => import('./pages/ForgotPasswordPage'));
const ConfirmForgotPasswordPage = lazy(() => import('./pages/ConfirmForgotPasswordPage'));
const AcceptInvitationPage = lazy(() => import('./pages/AcceptInvitationPage'));
const WelcomePage = lazy(() => import('./pages/WelcomePage'));

// 認証必要ページ（コア機能のみ。
// 削除済 (PR-A): PracticePage / ScoreHistoryPage / FavoritesPage / RankingPage /
//   TemplatePage / ReminderPage / SharedSessionsPage / WeeklyChallengePage /
//   AdminScenariosPage）
const MenuPage = lazy(() => import('./pages/MenuPage'));
const ProfilePage = lazy(() => import('./pages/ProfilePage'));
const AskAiPage = lazy(() => import('./pages/AskAiPage'));
const NotesPage = lazy(() => import('./pages/NotesPage'));
const NotificationPage = lazy(() => import('./pages/NotificationPage'));
const LearningReportPage = lazy(() => import('./pages/LearningReportPage'));
const HelpPage = lazy(() => import('./pages/HelpPage'));
const AdminInvitationsPage = lazy(() => import('./pages/AdminInvitationsPage'));
const AdminCompaniesPage = lazy(() => import('./pages/AdminCompaniesPage'));
const CodeEditorPage = lazy(() => import('./pages/CodeEditorPage'));

function NavigationToast() {
  const location = useLocation();
  const { showToast } = useToast();

  useEffect(() => {
    const toast = (location.state as { toast?: string })?.toast;
    if (toast) {
      showToast('success', toast);
      window.history.replaceState({}, '');
    }
  }, [location, showToast]);

  return null;
}

export default function App() {
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
        <Route path="/code-editor" element={<CodeEditorPage />} />
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

