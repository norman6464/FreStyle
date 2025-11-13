import { Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import LoginCallback from './pages/LoginCallback';
import HomePage from './components/HomePage';
import ChatPage from './pages/ChatPage';
import MenuPage from './pages/MenuPage';
import AskAiPage from './pages/AskAiPage';
import ConfirmPage from './pages/ConfirmPage';
import MemberPage from './pages/MemberPage';
import AddUserPage from './pages/AddUserPage';
import ProfilePage from './pages/ProfilePage';
import ForgotPasswordPage from './pages/ForgotPasswordPage';
import ConfirmForgotPasswordPage from './pages/ConfirmForgotPasswordPage';

export default function App() {
  return (
    <Routes>
      {/* 誰でもアクセス可能なルート */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/login/callback" element={<LoginCallback />} />
      <Route path="/signup" element={<SignupPage />} />
      <Route path="/confirm" element={<ConfirmPage />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route
        path="/confirm-forgot-password"
        element={<ConfirmForgotPasswordPage />}
      />

      {/* 認証が必要なルート */}
      <Route path="/" element={<MenuPage />} />
      <Route path="/profile/me" element={<ProfilePage />} />
      <Route path="/chat/members" element={<MemberPage />} />
      <Route path="/chat/users" element={<AddUserPage />} />
      <Route path="/chat/users/:roomId" element={<ChatPage />} />
      <Route path="/chat/ask-ai" element={<AskAiPage />} />

      {/* 404 */}
      <Route
        path="*"
        element={
          <div className="text-center mt-20 text-xl">
            ページが見つかりません
          </div>
        }
      />
    </Routes>
  );
}
