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
