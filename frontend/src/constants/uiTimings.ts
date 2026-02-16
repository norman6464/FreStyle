/** UI共通のタイミング・ディメンション定数 */

export const UI_TIMINGS = {
  /** コピー完了フィードバックの表示時間 (ms) */
  COPY_FEEDBACK_DURATION: 2000,
  /** ツールバーのblur非表示遅延 (ms) */
  TOOLBAR_HIDE_DELAY: 150,
  /** ブロック挿入ボタンの非表示遅延 (ms) */
  INSERTER_HIDE_DELAY: 200,
  /** マウス移動のスロットル間隔 (ms) */
  MOUSE_MOVE_THROTTLE: 50,
} as const;

export const UI_DIMENSIONS = {
  /** 選択ツールバーの上オフセット (px) */
  SELECTION_TOOLBAR_OFFSET_TOP: 48,
} as const;
