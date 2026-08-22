import { useEffect, useRef } from 'react';
import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { Placeholder } from '@tiptap/extensions';
import RichTextEditorToolbar from './RichTextEditorToolbar';
import SaveStatusIndicator, { type SaveStatus } from './SaveStatusIndicator';
import type { RichDocContent } from './emptyRichDoc';
import './richTextEditor.css';

export interface RichTextEditorProps {
  /** 表示・編集する tiptap ドキュメント JSON（正本）。 */
  value: RichDocContent;
  /** 本文が変わるたびに最新の doc JSON を返す（editable=false のときは呼ばれない）。 */
  onChange?: (value: RichDocContent) => void;
  /** 編集可否。false で読み取り専用（ツールバー非表示）。既定 true。 */
  editable?: boolean;
  /** 空のときに表示するプレースホルダ。 */
  placeholder?: string;
  /** 編集領域のアクセシブルネーム。 */
  ariaLabel?: string;
  /** 保存状態の表示（未指定なら表示しない）。保存の実処理は画面側が持つ。 */
  saveStatus?: SaveStatus;
  /** 外枠に付与する追加クラス。 */
  className?: string;
}

/**
 * RichTextEditor は tiptap ベースのリッチテキストエディタ（土台）。
 * doc JSON を value/onChange で制御する。書式は StarterKit の基本ノード/マークに絞り、
 * 画像アップロード・スラッシュコマンド・構文ハイライトは後続 PR で足す。
 *
 * ビジネスを知らない再利用資産（shared/ui）として置く。保存フロー（debounce・PUT・楽観ロック）は
 * この部品ではなく利用側の画面が担う。
 */
export default function RichTextEditor({
  value,
  onChange,
  editable = true,
  placeholder = '本文を入力…',
  ariaLabel = '本文',
  saveStatus,
  className = '',
}: RichTextEditorProps) {
  // onChange は props で差し替わり得るので ref 越しに最新を呼ぶ（onUpdate クロージャの陳腐化を防ぐ）。
  const onChangeRef = useRef(onChange);
  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  // 直近で「これが現在値」とみなしている doc の文字列表現。内容ベースで重複 emit を弾く。
  // tiptap はマウント時に一度 onUpdate を発火する（内容は初期 value と同じ）ため、初期値で初期化して
  // その空振り emit を握りつぶす（読み込み直後に「未保存」へ落ちないようにする）。
  const lastValueRef = useRef(JSON.stringify(value));

  const editor = useEditor({
    editable,
    extensions: [
      StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
      Placeholder.configure({ placeholder }),
    ],
    content: value,
    editorProps: {
      attributes: {
        class: 'focus:outline-none',
        role: 'textbox',
        'aria-multiline': 'true',
        'aria-label': ariaLabel,
      },
    },
    onUpdate: ({ editor: currentEditor }) => {
      const next = currentEditor.getJSON() as RichDocContent;
      const nextStr = JSON.stringify(next);
      // 内容が現在値と同じ（マウント時の空振り or 外部同期のエコー）なら通知しない。
      if (nextStr === lastValueRef.current) return;
      lastValueRef.current = nextStr;
      onChangeRef.current?.(next);
    },
  });

  // 外部から value が差し替わったとき（別ドキュメント読み込み等）だけ内容を同期する。
  // 自分の編集で親が value を更新した場合は現在値と一致するので setContent しない
  // （＝キャレットが飛ばず、無限ループにもならない）。
  useEffect(() => {
    if (!editor) return;
    const valueStr = JSON.stringify(value);
    if (valueStr !== lastValueRef.current) {
      lastValueRef.current = valueStr;
      editor.commands.setContent(value, { emitUpdate: false });
    }
  }, [editor, value]);

  // editable の変更を反映する。
  useEffect(() => {
    editor?.setEditable(editable);
  }, [editor, editable]);

  return (
    <div
      className={`overflow-hidden rounded-lg border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] ${className}`}
    >
      {editable && editor && <RichTextEditorToolbar editor={editor} />}
      <div className="rte-content prose prose-sm max-w-none px-4">
        <EditorContent editor={editor} />
      </div>
      {saveStatus && saveStatus !== 'idle' && (
        <div className="flex justify-end border-t border-[var(--color-surface-3)] px-4 py-1.5">
          <SaveStatusIndicator status={saveStatus} />
        </div>
      )}
    </div>
  );
}
