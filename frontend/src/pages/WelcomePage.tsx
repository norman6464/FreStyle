import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import AuthLayout from '../components/AuthLayout';
import PrimaryButton from '../components/PrimaryButton';
import authRepository, { UserInfo } from '../repositories/AuthRepository';
import { markOnboarded } from '../store/authSlice';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import { logger } from '../lib/logger';

/**
 * Welcome / オンボーディング画面（初回ログイン直後にのみ表示）。
 *
 * 業界標準 (Slack / Notion / Linear) に倣った 1 画面 / 1 ボタン構成。
 * プロフィール詳細などはあえて聞かず、「ようこそ」と「できること」を簡単に示してから
 * 「はじめる」で onboarded_at を更新してホームに遷移させる。
 *
 * ガード:
 *   - 既に onboarded === true なら即 / にリダイレクト（二度通させない）
 *
 * 詳細設計: frestyle-pdm/docs/auth/auth-flow-screen-transitions.md §4
 */
export default function WelcomePage() {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const onboarded = useSelector((state: RootState) => state.auth.onboarded);

  const [me, setMe] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  // 既に Welcome 完了済みの人がブラウザバック等でこの画面に来たら即座にホームへ。
  useEffect(() => {
    if (onboarded) {
      navigate('/', { replace: true });
    }
  }, [onboarded, navigate]);

  useEffect(() => {
    let cancelled = false;
    authRepository
      .getCurrentUser()
      .then((user) => {
        if (cancelled) return;
        setMe(user);
      })
      .catch((err) => {
        logger.error(err);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const handleStart = async () => {
    setSubmitting(true);
    try {
      await authRepository.completeOnboarding();
      // store を即時更新して Protected の /welcome リダイレクトループを回避
      dispatch(markOnboarded());
      navigate('/', { replace: true, state: { toast: 'FreStyle へようこそ！' } });
    } catch (err) {
      logger.error(err);
      // 失敗してもブロックしない: store だけ更新してホームに進ませる（次回 /auth/me で
      // 同期される。重要操作ではないので UX を優先）。
      dispatch(markOnboarded());
      navigate('/', { replace: true });
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <Loading fullscreen message="準備中..." />;
  }

  const roleLabel = me?.role ? roleLabels[me.role] ?? '' : '';
  const greetName = me?.displayName || me?.name || '';

  return (
    <AuthLayout title="FreStyle へようこそ">
      <div className="space-y-5">
        <p className="text-[var(--color-text-muted)] leading-relaxed">
          {greetName ? `${greetName} さん、` : ''}
          ご登録ありがとうございます。
          {roleLabel && (
            <>
              <br />
              あなたは <strong className="text-[var(--color-text)]">{roleLabel}</strong>
              {' '}としてアカウントが作成されました。
            </>
          )}
        </p>

        <div className="bg-[var(--color-surface-alt)] rounded-lg p-4 text-sm">
          <h3 className="font-bold mb-2">FreStyle でできること</h3>
          <ul className="list-disc list-inside space-y-1 text-[var(--color-text-muted)]">
            <li>AI チャットでビジネスメールの添削や言い換え提案</li>
            <li>練習モードでビジネスシナリオのロールプレイ</li>
            <li>学習ノート / お気に入りフレーズ / 進捗の可視化</li>
            {(me?.role === 'company_admin' || me?.role === 'super_admin') && (
              <li>管理画面で自社メンバーの招待 / 管理</li>
            )}
          </ul>
        </div>

        <p className="text-xs text-[var(--color-text-muted)]">
          表示名やプロフィール画像は、ログイン後の「プロフィール」画面でいつでも変更できます。
        </p>

        <PrimaryButton type="button" onClick={handleStart} loading={submitting}>
          はじめる
        </PrimaryButton>
      </div>
    </AuthLayout>
  );
}

const roleLabels: Record<string, string> = {
  super_admin: '運営管理者',
  company_admin: '会社管理者',
  trainee: '受講者',
};
