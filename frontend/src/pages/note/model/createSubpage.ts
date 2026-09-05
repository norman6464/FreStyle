import {
  NOTE_NEW_PAGE_TITLE,
  NoteRepository,
  emitNoteTreeEvent,
  type NoteResolvedPage,
} from '@/entities/note';

/**
 * createSubpage が editor に求める最小の形。tiptap の Editor はこれを満たす。
 * 全体の型に依存しないのは、この関数の関心が「参照を 1 つ挿す」ことだけだから
 * （テストもこの形の偽物で書ける）。
 */
export interface SubpageEditor {
  chain(): {
    focus(): {
      insertContent(content: {
        type: 'pageRef';
        attrs: { pageId: string; title: string };
      }): { run(): void };
    };
  };
}

/**
 * createSubpage は /page コマンドの本体。いま開いているページの子ページを作り、
 * 本文のカーソル位置に**ページ参照**（題名がリンク先に追従するインラインの 1 要素）を
 * 挿し、開くべき URL を返す。
 *
 * 順序が要点: 先に作る（参照に要る ID はサーバーが発番する）→ 参照を挿す
 * （エディタの onChange が発火して本文の自動保存が拾う）→ 木へ知らせる → 遷移は
 * 呼び出し側。参照の題名はサーバーが読み出しのたびに現在の値へ差し替えるので、
 * ここで入れる title は初回表示のための写しにすぎない。
 *
 * **失敗は投げる**（作れなかったのに参照だけ残る、を防ぐため挿入より前で落ちる）。
 */
export async function createSubpage(
  editor: SubpageEditor,
  resolved: NoteResolvedPage,
): Promise<string> {
  const child = await NoteRepository.createPage(resolved.workspaceSlug, resolved.page.spaceId, {
    title: NOTE_NEW_PAGE_TITLE,
    parentId: resolved.page.id,
  });
  editor
    .chain()
    .focus()
    .insertContent({
      type: 'pageRef',
      attrs: { pageId: child.id, title: child.title },
    })
    .run();
  emitNoteTreeEvent({ type: 'page-created', page: child });
  return `/kb/${child.id}`;
}
