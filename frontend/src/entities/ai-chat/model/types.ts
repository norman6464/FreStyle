/**
 * AI チャット（ai-chat）entity のドメイン型。
 */

/**
 * AiSession は AskAi 画面表示用の view 型。userId / updatedAt は不要なケースが多く、
 * scene のような UI 由来のフィールドを足してある。後段で AiChatSession へ統一予定。
 */
export interface AiSession {
  id: number;
  title?: string;
  scene?: string;
  sessionType?: string;
  scenarioId?: number;
  createdAt?: string;
}

/** AiAttachmentKind は添付が画像かドキュメントかの区分（表示の出し分けに使う）。 */
export type AiAttachmentKind = 'image' | 'document';

/**
 * AiAttachmentFormat は AWS Bedrock Converse の image/document format に渡す短い文字列。
 * バリデーションは backend 側 (`usecase.AllowedAttachmentContentTypes`) が一次情報。
 */
export type AiAttachmentFormat =
  | 'png'
  | 'jpeg'
  | 'gif'
  | 'webp'
  | 'pdf'
  | 'csv'
  | 'txt';

/**
 * AiAttachment は AI チャットメッセージに添付された画像 / ドキュメントの参照。
 * 実体は S3 にあり `key` で一意特定する。`previewUrl` は送信前のローカルプレビュー
 * （Object URL）を保持する用途で、バックエンドへは送らない。
 */
export interface AiAttachment {
  key: string;
  filename: string;
  contentType: string;
  kind: AiAttachmentKind;
  format: AiAttachmentFormat;
  sizeBytes: number;
  /** ローカル状態用: ブラウザ内 preview URL（送信完了後は破棄） */
  previewUrl?: string;
}

/**
 * AiMessage は AskAi 画面表示用の view 型。
 * `id` (string) を画面側で扱いやすい一意キーとして使う。
 * `createdAt` は ISO 8601 文字列（backend の SSE / REST 双方が ISO で返す）。
 */
export interface AiMessage {
  id: string;
  sessionId: number;
  content: string;
  role: 'user' | 'assistant';
  attachments?: AiAttachment[];
  createdAt?: string;
  isSender?: boolean;
  isDeleted?: boolean;
  /**
   * ストリーミング placeholder 由来のクライアント側 ID(FRESTYLE-146)。
   * done で id がサーバ確定値に差し替わっても React の key をこれで安定させ、
   * バブルの remount(ペーシングの残り放出が全文ジャンプになる)を防ぐ。
   */
  clientId?: string;
}
