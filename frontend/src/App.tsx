import { Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import LoginCallback from './pages/LoginCallback';
import ChatPage from './pages/ChatPage';
import ChatListPage from './pages/ChatListPage';
import MenuPage from './pages/MenuPage';
import AskAiPage from './pages/AskAiPage';
import PracticePage from './pages/PracticePage';
import ScoreHistoryPage from './pages/ScoreHistoryPage';
import FavoritesPage from './pages/FavoritesPage';
import ConfirmPage from './pages/ConfirmPage';
import MemberPage from './pages/MemberPage';
import AddUserPage from './pages/AddUserPage';
import ProfilePage from './pages/ProfilePage';
import UserProfilePage from './pages/UserProfilePage';
import NotesPage from './pages/NotesPage';
import NotificationPage from './pages/NotificationPage';
import LearningReportPage from './pages/LearningReportPage';
import ForgotPasswordPage from './pages/ForgotPasswordPage';
import ConfirmForgotPasswordPage from './pages/ConfirmForgotPasswordPage';
import AuthInitializer from './utils/AuthInitializer';
import Protected from './utils/Protected';
import AppShell from './components/layout/AppShell';
import ErrorBoundary from './components/ErrorBoundary';
import { ToastProvider } from './hooks/useToast';
import ToastContainer from './components/ToastContainer';


export default function App() {
  return (
    <ErrorBoundary>
    <ToastProvider>
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
        <Route path="/profile/personality" element={<UserProfilePage />} />
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
      </Route>
    </Routes>
    <ToastContainer />
    </ToastProvider>
    </ErrorBoundary>
  );
}
