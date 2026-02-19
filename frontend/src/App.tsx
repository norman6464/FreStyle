import { useEffect, lazy, Suspense } from 'react';
import { Routes, Route, useLocation } from 'react-router-dom';
import AuthInitializer from './utils/AuthInitializer';
import Protected from './utils/Protected';
import AppShell from './components/layout/AppShell';
import ErrorBoundary from './components/ErrorBoundary';
import Loading from './components/Loading';
import { ToastProvider, useToast } from './hooks/useToast';
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
const MemberPage = lazy(() => import('./pages/MemberPage'));
const ChatListPage = lazy(() => import('./pages/ChatListPage'));
const AddUserPage = lazy(() => import('./pages/AddUserPage'));
const ChatPage = lazy(() => import('./pages/ChatPage'));
const PracticePage = lazy(() => import('./pages/PracticePage'));
const ScoreHistoryPage = lazy(() => import('./pages/ScoreHistoryPage'));
const FavoritesPage = lazy(() => import('./pages/FavoritesPage'));
const AskAiPage = lazy(() => import('./pages/AskAiPage'));
const NotesPage = lazy(() => import('./pages/NotesPage'));
const NotificationPage = lazy(() => import('./pages/NotificationPage'));
const LearningReportPage = lazy(() => import('./pages/LearningReportPage'));
const FriendshipPage = lazy(() => import('./pages/FriendshipPage'));

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
        <Route path="/chat/members" element={<MemberPage />} />
        <Route path="/chat" element={<ChatListPage />} />
        <Route path="/chat/users" element={<AddUserPage />} />
        <Route path="/chat/users/:roomId" element={<ChatPage />} />
        <Route path="/practice" element={<PracticePage />} />
        <Route path="/scores" element={<ScoreHistoryPage />} />
        <Route path="/favorites" element={<FavoritesPage />} />
        <Route path="/chat/ask-ai" element={<AskAiPage />} />
        <Route path="/chat/ask-ai/:sessionId" element={<AskAiPage />} />
        <Route path="/notes" element={<NotesPage />} />
        <Route path="/notifications" element={<NotificationPage />} />
        <Route path="/reports" element={<LearningReportPage />} />
        <Route path="/friends" element={<FriendshipPage />} />
      </Route>
    </Routes>
    </Suspense>
    <NavigationToast />
    <ToastContainer />
    </ToastProvider>
    </ErrorBoundary>
  );
}
