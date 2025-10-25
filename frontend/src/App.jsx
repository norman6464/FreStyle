import { Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import LoginCallback from './pages/LoginCallback';
import HomePage from './components/HomePage';
import { useSelector } from 'react-redux';
import ChatPage from './pages/ChatPage';
import MenuPage from './pages/MenuPage';
import AskAiPage from './pages/AskAiPage';
import ConfirmPage from './pages/ConfirmPage';
import MemberPage from './pages/MemberPage';
import AddUserPage from './pages/AddUserPage';
import ProfilePage from './pages/ProfilePage';
export default function App() {
  const accessToken = useSelector((state) => state.auth.accessToken);

  return (
    <Routes>
      <Route
        path="/"
        element={accessToken ? <MenuPage /> : <Navigate to="/login" />}
      />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/login/callback" element={<LoginCallback />} />
      <Route path="/profile/me" element={<ProfilePage />} />
      <Route path="/chat/members" element={<MemberPage />} />
      <Route path="/chat/users" element={<AddUserPage />} />
      <Route path="/chat/users/:roomId" element={<ChatPage />} />
      <Route path="/chat/ask-ai" element={<AskAiPage />} />
      <Route path="/signup" element={<SignupPage />} />
      <Route path="/confirm" element={<ConfirmPage />} />
      {/* 404 redirect */}
      <Route
        path="*"
        element={
          <div className="text-center mt-20 text-xl">
            ページが見つかりません1
          </div>
        }
      />
    </Routes>
  );
}
