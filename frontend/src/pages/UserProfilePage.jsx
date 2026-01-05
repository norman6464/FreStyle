import { useState, useEffect } from 'react';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import HamburgerMenu from '../components/HamburgerMenu';
import { clearAuth } from '../store/authSlice';

export default function UserProfilePage() {
  const [form, setForm] = useState({
    displayName: '',
    selfIntroduction: '',
    communicationStyle: '',
    personalityTraits: [],
    goals: '',
    concerns: '',
    preferredFeedbackStyle: '',
  });
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isNewProfile, setIsNewProfile] = useState(false);
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // コミュニケーションスタイルの選択肢
  const communicationStyles = [
    { value: '', label: '選択してください' },
    { value: 'casual', label: 'カジュアル' },
    { value: 'formal', label: 'フォーマル' },
    { value: 'friendly', label: 'フレンドリー' },
    { value: 'professional', label: 'プロフェッショナル' },
  ];

  // 性格特性の選択肢
  const personalityOptions = [
    '内向的',
    '外向的',
    '論理的',
    '感情的',
    '共感力が高い',
    '分析的',
    'クリエイティブ',
    '計画的',
    '柔軟性がある',
    'リーダーシップがある',
  ];

  // フィードバックスタイルの選択肢
  const feedbackStyles = [
    { value: '', label: '選択してください' },
    { value: 'direct', label: 'ストレート（はっきり伝えてほしい）' },
    { value: 'gentle', label: 'やさしく（配慮を持って伝えてほしい）' },
    { value: 'detailed', label: '詳細に（具体的に説明してほしい）' },
  ];

  // ----------------------------
  // プロフィール取得
  // ----------------------------
  const fetchProfile = async () => {
    try {
      console.log('[UserProfilePage] Fetching user profile');
      const res = await fetch(`${API_BASE_URL}/api/user-profile/me`, {
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      // トークン期限切れならリフレッシュ
      if (res.status === 401) {
        console.warn('[UserProfilePage] Access token expired, attempting refresh');
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!refreshRes.ok) {
          console.error('[UserProfilePage] ERROR: Token refresh failed');
          dispatch(clearAuth());
          return;
        }

        const retryRes = await fetch(`${API_BASE_URL}/api/user-profile/me`, {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });

        const retryData = await retryRes.json();
        if (retryData.message) {
          // プロファイル未設定
          setIsNewProfile(true);
        } else {
          setForm({
            displayName: retryData.displayName || '',
            selfIntroduction: retryData.selfIntroduction || '',
            communicationStyle: retryData.communicationStyle || '',
            personalityTraits: retryData.personalityTraits || [],
            goals: retryData.goals || '',
            concerns: retryData.concerns || '',
            preferredFeedbackStyle: retryData.preferredFeedbackStyle || '',
          });
        }
        setLoading(false);
        return;
      }

      const data = await res.json();
      if (data.message) {
        // プロファイル未設定
        setIsNewProfile(true);
      } else {
        setForm({
          displayName: data.displayName || '',
          selfIntroduction: data.selfIntroduction || '',
          communicationStyle: data.communicationStyle || '',
          personalityTraits: data.personalityTraits || [],
          goals: data.goals || '',
          concerns: data.concerns || '',
          preferredFeedbackStyle: data.preferredFeedbackStyle || '',
        });
      }
    } catch (err) {
      console.error('[UserProfilePage] ERROR:', err.message);
      setMessage({ type: 'error', text: err.message });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  // ----------------------------
  // 性格特性のトグル
  // ----------------------------
  const togglePersonalityTrait = (trait) => {
    setForm((prev) => {
      const traits = prev.personalityTraits.includes(trait)
        ? prev.personalityTraits.filter((t) => t !== trait)
        : [...prev.personalityTraits, trait];
      return { ...prev, personalityTraits: traits };
    });
  };

  // ----------------------------
  // プロフィール保存（upsert）
  // ----------------------------
  const handleSave = async (e) => {
    e.preventDefault();
    try {
      console.log('[UserProfilePage] Saving user profile');
      const res = await fetch(`${API_BASE_URL}/api/user-profile/me/upsert`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(form),
      });

      // 401 → トークン更新
      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!refreshRes.ok) {
          navigate('/login');
          return;
        }

        const retryRes = await fetch(`${API_BASE_URL}/api/user-profile/me/upsert`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(form),
        });

        const retryData = await retryRes.json();
        if (!retryRes.ok) {
          throw new Error(retryData.error || '保存に失敗しました。');
        }

        setMessage({ type: 'success', text: 'パーソナリティ設定を保存しました。' });
        setIsNewProfile(false);
        return;
      }

      const data = await res.json();
      
      // 既に存在する場合は更新APIを使用
      if (res.status === 400 && data.error?.includes('既に存在')) {
        const updateRes = await fetch(`${API_BASE_URL}/api/user-profile/me/upsert`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(form),
        });

        const updateData = await updateRes.json();
        if (!updateRes.ok) {
          throw new Error(updateData.error || '更新に失敗しました。');
        }

        setMessage({ type: 'success', text: 'パーソナリティ設定を更新しました。' });
        return;
      }

      if (!res.ok) {
        throw new Error(data.error || '保存に失敗しました。');
      }

      setMessage({ type: 'success', text: 'パーソナリティ設定を保存しました。' });
      setIsNewProfile(false);
    } catch (error) {
      console.error('[UserProfilePage] ERROR:', error.message);
      setMessage({ type: 'error', text: error.message || '通信エラーが発生しました。' });
    }
  };

  // ローディング時
  if (loading) {
    return (
      <>
        <HamburgerMenu title="パーソナリティ設定" />
        <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 flex items-center justify-center pt-20">
          <div className="text-center">
            <div className="animate-pulse">
              <div className="w-16 h-16 bg-primary-200 rounded-full mx-auto mb-4"></div>
              <p className="text-gray-600">読み込み中...</p>
            </div>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <HamburgerMenu title="パーソナリティ設定" />
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 pt-20 pb-8 px-4">
        <div className="max-w-2xl mx-auto">
          {/* メッセージ */}
          {message && (
            <div
              className={`mb-6 p-4 rounded-lg border-l-4 flex items-start animate-fade-in ${
                message.type === 'error'
                  ? 'bg-red-50 border-red-500'
                  : 'bg-green-50 border-green-500'
              }`}
            >
              <div
                className={`flex-shrink-0 mr-3 ${
                  message.type === 'error' ? 'text-red-600' : 'text-green-600'
                }`}
              >
                {message.type === 'error' ? (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                      clipRule="evenodd"
                    />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                )}
              </div>
              <p
                className={
                  message.type === 'error'
                    ? 'text-red-700 font-semibold'
                    : 'text-green-700 font-semibold'
                }
              >
                {message.text}
              </p>
            </div>
          )}

          {/* メインカード */}
          <div className="bg-white rounded-2xl shadow-lg p-8 border border-gray-100">
            <div className="text-center mb-10">
              <div className="w-24 h-24 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full mx-auto flex items-center justify-center mb-4 shadow-lg">
                <svg
                  className="w-12 h-12 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                  />
                </svg>
              </div>
              <h2 className="text-3xl font-bold text-gray-800">
                {isNewProfile ? 'パーソナリティを設定' : 'パーソナリティを編集'}
              </h2>
              <p className="text-gray-600 mt-2">
                AIがあなたに合ったフィードバックを提供するための情報を設定してください
              </p>
            </div>

            <form onSubmit={handleSave} className="space-y-6">
              {/* 基本情報セクション */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-gray-800 border-b pb-2">
                  📝 基本情報
                </h3>
                <InputField
                  label="呼ばれたい名前"
                  name="displayName"
                  value={form.displayName}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, displayName: e.target.value }))
                  }
                  placeholder="例：タロウ、たろちゃん"
                />
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    自己紹介
                  </label>
                  <textarea
                    name="selfIntroduction"
                    value={form.selfIntroduction}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, selfIntroduction: e.target.value }))
                    }
                    placeholder="あなた自身について自由に書いてください..."
                    rows="3"
                    className="w-full border-2 border-gray-200 rounded-lg px-4 py-3 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 transition-all duration-200 resize-none"
                  />
                </div>
              </div>

              {/* コミュニケーションスタイル */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-gray-800 border-b pb-2">
                  💬 コミュニケーションスタイル
                </h3>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    あなたのコミュニケーションスタイル
                  </label>
                  <select
                    value={form.communicationStyle}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, communicationStyle: e.target.value }))
                    }
                    className="w-full border-2 border-gray-200 rounded-lg px-4 py-3 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 transition-all duration-200"
                  >
                    {communicationStyles.map((style) => (
                      <option key={style.value} value={style.value}>
                        {style.label}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-3">
                    性格特性（当てはまるものを選んでください）
                  </label>
                  <div className="flex flex-wrap gap-2">
                    {personalityOptions.map((trait) => (
                      <button
                        key={trait}
                        type="button"
                        onClick={() => togglePersonalityTrait(trait)}
                        className={`px-4 py-2 rounded-full text-sm font-medium transition-all duration-200 ${
                          form.personalityTraits.includes(trait)
                            ? 'bg-primary-500 text-white shadow-md'
                            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                        }`}
                      >
                        {trait}
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              {/* AIフィードバック設定 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-gray-800 border-b pb-2">
                  🤖 AIフィードバック設定
                </h3>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    コミュニケーションで改善したい点・目標
                  </label>
                  <textarea
                    name="goals"
                    value={form.goals}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, goals: e.target.value }))
                    }
                    placeholder="例：もっと簡潔に伝えられるようになりたい、相手の気持ちを考えた発言ができるようになりたい..."
                    rows="3"
                    className="w-full border-2 border-gray-200 rounded-lg px-4 py-3 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 transition-all duration-200 resize-none"
                  />
                </div>

                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    苦手なこと・気になっていること
                  </label>
                  <textarea
                    name="concerns"
                    value={form.concerns}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, concerns: e.target.value }))
                    }
                    placeholder="例：話が長くなりがち、相手の反応が気になる..."
                    rows="3"
                    className="w-full border-2 border-gray-200 rounded-lg px-4 py-3 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 transition-all duration-200 resize-none"
                  />
                </div>

                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    フィードバックの受け取り方
                  </label>
                  <select
                    value={form.preferredFeedbackStyle}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, preferredFeedbackStyle: e.target.value }))
                    }
                    className="w-full border-2 border-gray-200 rounded-lg px-4 py-3 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 transition-all duration-200"
                  >
                    {feedbackStyles.map((style) => (
                      <option key={style.value} value={style.value}>
                        {style.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <PrimaryButton type="submit">
                {isNewProfile ? 'パーソナリティを保存' : 'パーソナリティを更新'}
              </PrimaryButton>
            </form>
          </div>
        </div>
      </div>
    </>
  );
}
