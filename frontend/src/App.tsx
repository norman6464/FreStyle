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

// 認証必要ページ
const MenuPage = lazy(() => import('./pages/MenuPage'));
const ProfilePage = lazy(() => import('./pages/ProfilePage'));
const PracticePage = lazy(() => import('./pages/PracticePage'));
const ScoreHistoryPage = lazy(() => import('./pages/ScoreHistoryPage'));
const FavoritesPage = lazy(() => import('./pages/FavoritesPage'));
const AskAiPage = lazy(() => import('./pages/AskAiPage'));
const NotesPage = lazy(() => import('./pages/NotesPage'));
const NotificationPage = lazy(() => import('./pages/NotificationPage'));
const LearningReportPage = lazy(() => import('./pages/LearningReportPage'));
const RankingPage = lazy(() => import('./pages/RankingPage'));
const TemplatePage = lazy(() => import('./pages/TemplatePage'));
const ReminderPage = lazy(() => import('./pages/ReminderPage'));
const SharedSessionsPage = lazy(() => import('./pages/SharedSessionsPage'));
const WeeklyChallengePage = lazy(() => import('./pages/WeeklyChallengePage'));
const HelpPage = lazy(() => import('./pages/HelpPage'));
const AdminScenariosPage = lazy(() => import('./pages/AdminScenariosPage'));
const AdminInvitationsPage = lazy(() => import('./pages/AdminInvitationsPage'));
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

      {/* 認証が必要（AppShellレイアウト内） */}
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
        <Route path="/practice" element={<PracticePage />} />
        <Route path="/scores" element={<ScoreHistoryPage />} />
        <Route path="/favorites" element={<FavoritesPage />} />
        <Route path="/chat/ask-ai" element={<AskAiPage />} />
        <Route path="/chat/ask-ai/:sessionId" element={<AskAiPage />} />
        <Route path="/notes" element={<NotesPage />} />
        <Route path="/notifications" element={<NotificationPage />} />
        <Route path="/reports" element={<LearningReportPage />} />
        <Route path="/ranking" element={<RankingPage />} />
        <Route path="/templates" element={<TemplatePage />} />
        <Route path="/reminder" element={<ReminderPage />} />
        <Route path="/shared-sessions" element={<SharedSessionsPage />} />
        <Route path="/weekly-challenge" element={<WeeklyChallengePage />} />
        <Route path="/help" element={<HelpPage />} />
        {/* Admin 専用（コンポーネント側で isAdmin チェック → 非 admin は / にリダイレクト） */}
        <Route path="/code-editor" element={<CodeEditorPage />} />
        {/* Admin 専用（コンポーネント側で isAdmin チェック → 非 admin は / にリダイレクト） */}
        <Route path="/admin/scenarios" element={<AdminScenariosPage />} />
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
