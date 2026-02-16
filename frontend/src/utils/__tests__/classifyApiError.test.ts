import { describe, it, expect } from 'vitest';
import { AxiosError, AxiosHeaders } from 'axios';
import { classifyApiError } from '../classifyApiError';

function createAxiosError(status: number, message = 'Request failed'): AxiosError {
  const headers = new AxiosHeaders();
  return new AxiosError(message, 'ERR_BAD_REQUEST', undefined, undefined, {
    status,
    statusText: '',
    headers: {},
    config: { headers },
    data: {},
  });
}

function createNetworkError(): AxiosError {
  return new AxiosError('Network Error', 'ERR_NETWORK', undefined, undefined, undefined);
}

function createTimeoutError(): AxiosError {
  return new AxiosError('timeout of 5000ms exceeded', 'ECONNABORTED', undefined, undefined, undefined);
}

describe('classifyApiError', () => {
  it('403エラーでアクセス権限メッセージを返す', () => {
    const error = createAxiosError(403);
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('この操作を行う権限がありません。');
  });

  it('404エラーでリソース不在メッセージを返す', () => {
    const error = createAxiosError(404);
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('データが見つかりません。');
  });

  it('429エラーでレート制限メッセージを返す', () => {
    const error = createAxiosError(429);
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('リクエストが多すぎます。しばらくしてからお試しください。');
  });

  it('500エラーでサーバーエラーメッセージを返す', () => {
    const error = createAxiosError(500);
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('サーバーエラーが発生しました。時間をおいてお試しください。');
  });

  it('503エラーでサーバーエラーメッセージを返す', () => {
    const error = createAxiosError(503);
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('サーバーエラーが発生しました。時間をおいてお試しください。');
  });

  it('ネットワークエラーで接続エラーメッセージを返す', () => {
    const error = createNetworkError();
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('インターネット接続を確認してください。');
  });

  it('タイムアウトエラーでタイムアウトメッセージを返す', () => {
    const error = createTimeoutError();
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('リクエストがタイムアウトしました。再度お試しください。');
  });

  it('一般的なErrorでデフォルトメッセージを返す', () => {
    const error = new Error('something went wrong');
    const result = classifyApiError(error, 'セッション取得に失敗しました。');
    expect(result).toBe('セッション取得に失敗しました。');
  });

  it('不明なエラーでデフォルトメッセージを返す', () => {
    const result = classifyApiError('unknown', 'デフォルトメッセージ');
    expect(result).toBe('デフォルトメッセージ');
  });

  it('未対応のHTTPステータスコードでデフォルトメッセージを返す', () => {
    const error = createAxiosError(418);
    const result = classifyApiError(error, 'デフォルトメッセージ');
    expect(result).toBe('デフォルトメッセージ');
  });
});
