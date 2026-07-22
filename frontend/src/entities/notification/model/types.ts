/**
 * 通知（notification）entity のドメイン型。
 */

/** 通知（フロント表示用 view）。
 *  backend 1:1 は `NotificationDto` を参照すること。 */
export interface Notification {
  id: number;
  type: string;
  title: string;
  message: string;
  isRead: boolean;
  relatedId?: number;
  createdAt: string;
}
