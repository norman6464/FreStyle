import { Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import LoginCallback from './pages/LoginCallback';
import ChatPage from './pages/ChatPage';
import ChatListPage from './pages/ChatListPage';
import MenuPage from './pages/MenuPage';
import AskAiPage from './pages/AskAiPage';
import ConfirmPage from './pages/ConfirmPage';
import MemberPage from './pages/MemberPage';
import AddUserPage from './pages/AddUserPage';
import ProfilePage from './pages/ProfilePage';
import UserProfilePage from './pages/UserProfilePage';
import ForgotPasswordPage from './pages/ForgotPasswordPage';
import ConfirmForgotPasswordPage from './pages/ConfirmForgotPasswordPage';
import AuthInitializer from './utils/AuthInitializer';
import Protected from './utils/Protected';

export default function App() {
  return (
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

      {/* 認証が必要 */}
      <Route
        path="/"
        element={
          <AuthInitializer>
            <Protected>
              <MenuPage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/profile/me"
        element={
          <AuthInitializer>
            <Protected>
              <ProfilePage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/profile/personality"
        element={
          <AuthInitializer>
            <Protected>
              <UserProfilePage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/chat/members"
        element={
          <AuthInitializer>
            <Protected>
              <MemberPage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/chat"
        element={
          <AuthInitializer>
            <Protected>
              <ChatListPage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/chat/users"
        element={
          <AuthInitializer>
            <Protected>
              <AddUserPage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/chat/users/:roomId"
        element={
          <AuthInitializer>
            <Protected>
              <ChatPage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/chat/ask-ai"
        element={
          <AuthInitializer>
            <Protected>
              <AskAiPage />
            </Protected>
          </AuthInitializer>
        }
      />
      <Route
        path="/chat/ask-ai/:sessionId"
        element={
          <AuthInitializer>
            <Protected>
              <AskAiPage />
            </Protected>
          </AuthInitializer>
        }
      />
    </Routes>
  );
}
