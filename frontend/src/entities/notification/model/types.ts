/**
 * 通知（notification）entity のドメイン型。
 */

import type { components } from '@/generated/api';

/**
 * OpenAPI から生成した backend の通知スキーマ（`make openapi` で更新される）。
 * 手で書き写さず、下の互換チェックで食い違いを検出する。
 */
type NotificationSchema =
  components['schemas']['github_com_norman6464_FreStyle_backend_internal_domain.Notification'];

/**
 * 通知（フロント表示用 view）。
 *
 * 項目名は backend の JSON と一致させること。かつて本文を `message` として読んでいたため、
 * backend が `body` で返す本文が画面に出ず、企業申請通知の申請者名・会社名が
 * 見えない状態だった（FRESTYLE-87）。
 *
 * 生成スキーマは swaggo の仕様上すべて optional になるため、扱いやすさを優先して
 * ここでは必須で宣言し、代わりに下の型アサーションで生成スキーマとの整合を保証する。
 */
export interface Notification {
  id: number;
  type: string;
  title: string;
  body: string;
  isRead: boolean;
  createdAt: string;
}

/**
 * 生成スキーマとの互換チェック（実行時コストなし）。
 *
 * `Notification` に生成スキーマへ存在しない項目を足したり、項目名を変えたりすると
 * ここで型エラーになる。backend 側が項目名を変えた場合も `make openapi` 後にここで気づける。
 * 「送る側と受け取る側で名前がずれても誰も気づかない」状態を作らないための歯止め。
 */
type AssertNoUnknownField = keyof Notification extends keyof NotificationSchema ? true : never;
type AssertSameFieldTypes = Notification extends Required<Omit<NotificationSchema, 'userId'>>
  ? true
  : never;

// 使用しないと未使用型として lint に落ちるため、値として 1 度だけ参照する。
export const NOTIFICATION_CONTRACT_OK: AssertNoUnknownField & AssertSameFieldTypes = true;
