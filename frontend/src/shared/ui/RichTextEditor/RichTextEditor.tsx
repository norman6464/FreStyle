import { useCallback, useEffect, useRef } from 'react';
import { EditorContent, useEditor, type Editor } from '@tiptap/react';
import { createEditorExtensions } from './editorExtensions';
import { acceptedImageFiles, insertUploadedImages } from './imageInsertion';
import BubbleFormatMenu from './BubbleFormatMenu';
import SaveStatusIndicator, { type SaveStatus } from './SaveStatusIndicator';
import type { RichDocContent } from './emptyRichDoc';
import './richTextEditor.css';

export interface RichTextEditorProps {
  /** 表示・編集する tiptap ドキュメント JSON（正本）。 */
  value: RichDocContent;
  /** 本文が変わるたびに最新の doc JSON を返す（editable=false のときは呼ばれない）。 */
  onChange?: (value: RichDocContent) => void;
  /** 編集可否。false で読み取り専用（バブルメニュー非表示）。既定 true。 */
  editable?: boolean;
  /** 空のときに表示するプレースホルダ。 */
  placeholder?: string;
  /** 編集領域のアクセシブルネーム。 */
  ariaLabel?: string;
  /** 保存状態の表示（未指定なら表示しない）。保存の実処理は画面側が持つ。 */
  saveStatus?: SaveStatus;
  /**
   * 画像をアップロードして表示用 URL を返す。指定したときだけ画像挿入
   * （ドラッグ&ドロップ・貼り付け）が有効になる。
   * 失敗時の通知（トースト等）は呼び出し側の方針に委ねる（ここでは握りつぶす）。
   */
  onImageUpload?: (file: File) => Promise<string>;
  /**
   * エディタ生成直後に一度だけ呼ばれるライフサイクルフック。
   * 生成直後にフォーカスしたい・外部から editor を参照して拡張したい、といった用途の拡張点。
   */
  onCreate?: (editor: Editor) => void;
  /** 外枠に付与する追加クラス。 */
  className?: string;
}

/**
 * RichTextEditor は tiptap ベースのリッチテキストエディタ。
 * doc JSON を value/onChange で制御する。書式は StarterKit の基本ノード/マーク＋画像に対応。
 *
 * 見た目は枠のないインライン文書（固定ツールバーは持たず、テキスト選択時に浮かぶバブルメニューで
 * 書式を出す）。スラッシュコマンド・ドラッグハンドルは後続 PR で足す。
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
  onImageUpload,
  onCreate,
  className = '',
}: RichTextEditorProps) {
  // onChange は props で差し替わり得るので ref 越しに最新を呼ぶ（onUpdate クロージャの陳腐化を防ぐ）。
  const onChangeRef = useRef(onChange);
  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  // onCreate も生成時クロージャの陳腐化を避けるため ref 越しに最新を呼ぶ。
  const onCreateRef = useRef(onCreate);
  useEffect(() => {
    onCreateRef.current = onCreate;
  }, [onCreate]);

  // 画像アップロード関数・editor 参照を ref で持ち、editorProps（paste/drop）から最新を参照する。
  const onImageUploadRef = useRef(onImageUpload);
  useEffect(() => {
    onImageUploadRef.current = onImageUpload;
  }, [onImageUpload]);
  const editorRef = useRef<Editor | null>(null);

  // アンマウント（別ノートへ切替）後にアップロードが完了しても挿入しないための番人。
  const mountedRef = useRef(true);
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  // クリップボード/ドロップから画像ファイルだけ取り出し、選択順どおりに順次アップロード挿入する。
  const handleImageFiles = useCallback((files: FileList | null | undefined): boolean => {
    const upload = onImageUploadRef.current;
    const currentEditor = editorRef.current;
    if (!upload || !currentEditor) return false;
    const images = acceptedImageFiles(files);
    if (images.length === 0) return false;
    void insertUploadedImages(currentEditor, images, upload, () => mountedRef.current);
    return true;
  }, []);

  // 直近で「これが現在値」とみなしている doc の文字列表現。内容ベースで重複 emit を弾く。
  // tiptap はマウント時に一度 onUpdate を発火する（内容は初期 value と同じ）ため、初期値で初期化して
  // その空振り emit を握りつぶす（読み込み直後に「未保存」へ落ちないようにする）。
  const lastValueRef = useRef(JSON.stringify(value));

  const editor = useEditor({
    editable,
    extensions: createEditorExtensions({ placeholder }),
    content: value,
    editorProps: {
      attributes: {
        class: 'focus:outline-none',
        role: 'textbox',
        'aria-multiline': 'true',
        'aria-label': ariaLabel,
      },
      // クリップボード/ドロップに画像ファイルがあればアップロードして挿入する。
      handlePaste: (_view, event) => handleImageFiles(event.clipboardData?.files),
      handleDrop: (_view, event) => {
        if (event.dataTransfer?.files && handleImageFiles(event.dataTransfer.files)) {
          event.preventDefault();
          return true;
        }
        return false;
      },
    },
    onCreate: ({ editor: currentEditor }) => {
      onCreateRef.current?.(currentEditor);
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
  editorRef.current = editor;

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
    <div className={`rte-root ${className}`}>
      <div className="rte-content prose max-w-none">
        <EditorContent editor={editor} />
      </div>
      {editable && editor && <BubbleFormatMenu editor={editor} />}
      {saveStatus && saveStatus !== 'idle' && (
        <div className="mt-2 flex justify-end">
          <SaveStatusIndicator status={saveStatus} />
        </div>
      )}
    </div>
  );
}
