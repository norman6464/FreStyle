import type { KbPage, KbWorkspace } from './types';

/**
 * ページ画面からサイドバーの木へ「変わったよ」を知らせる通知。
 *
 * ページ本体（/p の画面）とサイドバーの木は別々の状態を持つ。画面側でページを
 * 作ったり題名を変えたりしたとき、木が知らないままだと表示が食い違う。
 * props で結ぶには両者が遠すぎる（木の状態は widget の hook の中にある）ので、
 * entities 層の小さな購読口で結ぶ。イベントはサーバーが返した確定後のページを運ぶ
 * （楽観更新の通知ではない — 失敗した操作がイベントになることはない）。
 *
 * ワークスペースの作成・削除も同じ理由で通知する。SecondaryPanel はモバイル用/
 * デスクトップ用の DOM を常に両方マウントするため、KbSidebar（useKbTree）は
 * 単一画面内でも複数インスタンスが独立に一覧を持つ。ヘッダーの切替（useWorkspaceList）
 * も別インスタンスなので、片方の変更を他方が自動では知れない。
 */
export type KbTreeEvent =
  | { type: 'page-created'; page: KbPage }
  | { type: 'page-renamed'; page: KbPage }
  /** 物理削除（子孫ごと消えた）。開いている画面が「消えた場所」かはページ側が判定する。 */
  | { type: 'page-deleted'; pageId: string }
  | { type: 'workspace-created'; workspace: KbWorkspace }
  /** 配下は FK CASCADE で全消去。開いている画面がその配下かはページ側が判定する。 */
  | { type: 'workspace-deleted'; workspaceSlug: string };

type KbTreeEventListener = (event: KbTreeEvent) => void;

const listeners = new Set<KbTreeEventListener>();

/** subscribeKbTreeEvents は購読を開始し、解除関数を返す。 */
export function subscribeKbTreeEvents(listener: KbTreeEventListener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

/** emitKbTreeEvent は購読者全員へ知らせる。購読者がいなければ何もしない。 */
export function emitKbTreeEvent(event: KbTreeEvent): void {
  // 途中で購読が外れても走査が壊れないよう写しを回す。
  for (const listener of [...listeners]) {
    listener(event);
  }
}
