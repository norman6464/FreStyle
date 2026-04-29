/**
 * 軽量 logger ユーティリティ。
 *
 * 設計方針:
 * - production build (`import.meta.env.PROD === true`) では `debug` / `info` / `warn` を no-op に
 *   し、`error` のみ出力する（ブラウザ DevTools / CloudWatch RUM 等のエラー観測経路に残す）
 * - development では従来どおり全レベル出力
 * - production code 全体で `console.*` を直接使うのを禁じ、こちらを経由させることで
 *   将来 Sentry / Datadog RUM 等への送信に切り替えるときの単一切替点にする
 *
 * 使い方:
 *   import { logger } from '@/lib/logger';
 *   logger.error('プロフィール画像のアップロードに失敗しました:', error);
 *
 * 注: テストコード (vitest 配下) では console.* / vi.spyOn(console, ...) を直接使い続ける。
 * production-only な抽象なので test には侵入させない。
 */
const isProd = typeof import.meta !== 'undefined' && import.meta.env?.PROD === true;

type LogArgs = readonly unknown[];

const noop = (..._args: LogArgs): void => {
  // production build では何もしない
};

export const logger = {
  /** 開発時のみ出力（production では完全に無効） */
  debug: isProd ? noop : (...args: LogArgs) => console.debug(...args),
  info: isProd ? noop : (...args: LogArgs) => console.info(...args),
  warn: isProd ? noop : (...args: LogArgs) => console.warn(...args),
  /**
   * production / development 共に出力する。
   * 将来 Sentry / Datadog RUM への送信に拡張する際はここに forwarder を追加。
   */
  error: (...args: LogArgs) => console.error(...args),
};
