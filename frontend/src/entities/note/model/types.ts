/**
 * ノート（note）entity のドメイン型。
 *
 * SessionNote は AI チャットのセッションに紐づくノートだが、
 * 実体はノートの一種なので note Slice に置く。
 */

/** セッションメモ */
export interface SessionNote {
  sessionId: number;
  note: string;
  updatedAt: string;
}

/**
 * Note は Go backend `domain.Note` と 1:1 で対応する。
 * - id / userId は number
 * - createdAt / updatedAt は RFC3339 string（ISO）
 * - isPublic = 公開フラグ、isPinned = ピン留めフラグ（独立した属性）
 */
export interface Note {
  id: number;
  userId: number;
  title: string;
  content: string;
  isPublic: boolean;
  isPinned: boolean;
  createdAt: string;
  updatedAt: string;
}
