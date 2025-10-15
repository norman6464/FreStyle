import { Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/login" />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/signup" element={<SignupPage />} />
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
