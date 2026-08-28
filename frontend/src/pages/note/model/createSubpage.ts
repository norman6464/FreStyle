import {
  NOTE_NEW_PAGE_TITLE,
  NoteRepository,
  emitNoteTreeEvent,
  type NoteResolvedPage,
} from '@/entities/note';

/**
 * createSubpage が editor に求める最小の形。tiptap の Editor はこれを満たす。
 * 全体の型に依存しないのは、この関数の関心が「リンクを 1 つ挿す」ことだけだから
 * （テストもこの形の偽物で書ける）。
 */
export interface SubpageEditor {
  chain(): {
    focus(): {
      insertContent(content: {
        type: 'text';
        text: string;
        marks: { type: 'link'; attrs: { href: string } }[];
      }): { run(): void };
    };
  };
}

/**
 * createSubpage は /page コマンドの本体。いま開いているページの子ページを作り、
 * 本文のカーソル位置にそのページへのリンクを挿し、開くべき URL を返す。
 *
 * 順序が要点: 先に作る（リンクに要る ID はサーバーが発番する）→ リンクを挿す
 * （エディタの onChange が発火して本文の自動保存が拾う）→ 木へ知らせる → 遷移は
 * 呼び出し側。リンクの文字は作成時の題名のまま追従しない（プレーンなリンク）。
 *
 * **失敗は投げる**（作れなかったのにリンクだけ残る、を防ぐため挿入より前で落ちる）。
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
      type: 'text',
      text: child.title,
      marks: [{ type: 'link', attrs: { href: `/p/${child.id}` } }],
    })
    .run();
  emitNoteTreeEvent({ type: 'page-created', page: child });
  return `/p/${child.id}`;
}
