/*
 * entities/ai-chat の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { default as AiChatRepository } from './api/aiChatRepository';
export type { CreateSessionRequest } from './api/aiChatRepository';
export type { UpdateSessionTitleRequest } from './api/aiChatRepository';
export type { IssueAttachmentUploadUrlRequest } from './api/aiChatRepository';
export type { AttachmentUploadUrlResponse } from './api/aiChatRepository';

export type {
  AiSession,
  AiMessage,
  AiAttachment,
  AiAttachmentKind,
  AiAttachmentFormat,
} from './model/types';

export { default as AiSessionListItem } from './ui/AiSessionListItem';
export { default as MessageBubble } from './ui/MessageBubble';
export { default as MessageBubbleAi } from './ui/MessageBubbleAi';
export { default as MessageActionRow } from './ui/MessageActionRow';
export { default as MessageAttachmentList } from './ui/MessageAttachmentList';
