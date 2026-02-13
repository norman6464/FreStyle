import axios, { AxiosError, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

/**
 * Axiosインスタンス
 *
 * <p>役割:</p>
 * <ul>
 *   <li>API呼び出しの一元管理</li>
 *   <li>認証トークンの自動リフレッシュ</li>
 *   <li>エラーハンドリングの統一</li>
 * </ul>
 *
 * <p>インフラ層（Infrastructure Layer）:</p>
 * <ul>
 *   <li>外部APIとの通信を担当</li>
 *   <li>HTTPクライアントの設定</li>
 * </ul>
 */
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Cookie（JWT）を自動送信
});

/**
 * トークンリフレッシュ中フラグ
 * 複数のリクエストが同時に401を受けた場合、リフレッシュは1回だけ実行
 */
let isRefreshing = false;

/**
 * リフレッシュ待ちのリクエストキュー
 */
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

/**
 * キューの処理
 */
const processQueue = (error: AxiosError | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve();
    }
  });

  failedQueue = [];
};

/**
 * レスポンスインターセプター
 * 401エラー時に自動的にトークンをリフレッシュ
 */
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // 401エラーかつ、まだリトライしていない場合
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // 既にリフレッシュ中の場合はキューに追加
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(() => {
            return apiClient(originalRequest);
          })
          .catch((err) => {
            return Promise.reject(err);
          });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // トークンリフレッシュ
        await axios.post(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {},
          { withCredentials: true }
        );

        processQueue(null);
        isRefreshing = false;

        // リトライ
        return apiClient(originalRequest);
      } catch (refreshError) {
        processQueue(error);
        isRefreshing = false;

        // リフレッシュ失敗 → ログインページへ
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }

        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;
