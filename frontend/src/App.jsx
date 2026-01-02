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
import { useSelector } from 'react-redux';
import AuthInitializer from './utils/AuthInitializer';
import Protected from './utils/Protected';

export default function App() {
  
  const accessToken = useSelector((state) => state.auth.accessToken);
  
  return (
<AuthInitializer>
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
        <Route path="/" element={<Protected><MenuPage /></Protected>} />
        <Route path="/profile/me" element={<Protected><ProfilePage /></Protected>} />
        <Route path="/chat/members" element={<Protected><MemberPage /></Protected>} />
        <Route path="/chat/users" element={<Protected><AddUserPage /></Protected>} />
        <Route path="/chat/users/:roomId" element={<Protected><ChatPage /></Protected>} />
        <Route path="/chat/ask-ai" element={<Protected><AskAiPage /></Protected>} />

        {/* 404 */}
        <Route path="*" element={<NotFound />} />
      </Routes>
    </AuthInitializer>
  );
}
