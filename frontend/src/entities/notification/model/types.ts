/**
 * 通知（notification）entity のドメイン型。
 */

/**
 * 通知（フロント表示用 view）。
 *
 * 項目名は backend の JSON と一致させること。かつて本文を `message` として読んでいたため、
 * backend が `body` で返す本文が画面に出ず、企業申請通知の申請者名・会社名が
 * 見えない状態だった。
 */
export interface Notification {
  id: number;
  type: string;
  title: string;
  body: string;
  isRead: boolean;
  createdAt: string;
}
