import { AxiosError } from 'axios';

const STATUS_MESSAGES: Record<number, string> = {
  403: 'この操作を行う権限がありません。',
  404: 'データが見つかりません。',
  429: 'リクエストが多すぎます。しばらくしてからお試しください。',
};

function isServerError(status: number): boolean {
  return status >= 500;
}

export function classifyApiError(error: unknown, fallback: string): string {
  if (!(error instanceof AxiosError)) {
    return fallback;
  }

  if (error.code === 'ERR_NETWORK') {
    return 'インターネット接続を確認してください。';
  }

  if (error.code === 'ECONNABORTED') {
    return 'リクエストがタイムアウトしました。再度お試しください。';
  }

  const status = error.response?.status;
  if (!status) {
    return fallback;
  }

  if (STATUS_MESSAGES[status]) {
    return STATUS_MESSAGES[status];
  }

  if (isServerError(status)) {
    return 'サーバーエラーが発生しました。時間をおいてお試しください。';
  }

  return fallback;
}

export function extractServerErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof AxiosError) {
    const serverMessage = error.response?.data?.error;
    if (typeof serverMessage === 'string' && serverMessage) {
      return serverMessage;
    }
  }
  return classifyApiError(error, fallback);
}

/** API エラーから取り出した情報。axios 依存をこの module 内に閉じ込めるための型。 */
export interface ApiErrorInfo {
  /** HTTP ステータス（レスポンスがある場合）。 */
  status?: number;
  /** axios のエラーコード（ERR_NETWORK / ECONNABORTED など）。 */
  code?: string;
  /** backend が返す機械可読コード（{"error":"invitation_required"} の値）。 */
  serverCode?: string;
  /** backend が返す人間向けメッセージ（{"message":"..."} の値）。 */
  serverMessage?: string;
}

/**
 * getApiError は unknown なエラーから status / code / サーバ返却フィールドを取り出す。
 * hooks / pages が `axios.isAxiosError` を直接呼ばず、この helper 経由でステータスコード
 * 分岐できるようにする（axios への依存をこの module に集約する）。
 */
export function getApiError(error: unknown): ApiErrorInfo {
  if (error instanceof AxiosError) {
    const data = error.response?.data as { error?: string; message?: string } | undefined;
    return {
      status: error.response?.status,
      code: error.code,
      serverCode: data?.error,
      serverMessage: data?.message,
    };
  }
  return {};
}
