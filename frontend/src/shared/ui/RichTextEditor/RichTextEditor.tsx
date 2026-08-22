import { useCallback, useEffect, useRef } from 'react';
import { EditorContent, useEditor, type Editor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { Placeholder } from '@tiptap/extensions';
import Image from '@tiptap/extension-image';
import {
  ACCEPTED_IMAGE_ACCEPT_ATTR,
  isAcceptedImageMimeType,
} from '@/shared/config/imageUpload';
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
  /**
   * 画像をアップロードして表示用 URL を返す。指定したときだけ画像挿入 UI
   * （ツールバーの画像ボタン・ドラッグ&ドロップ・貼り付け）が有効になる。
   * 失敗時の通知（トースト等）は呼び出し側の方針に委ねる（ここでは握りつぶす）。
   */
  onImageUpload?: (file: File) => Promise<string>;
  /** 外枠に付与する追加クラス。 */
  className?: string;
}

/**
 * RichTextEditor は tiptap ベースのリッチテキストエディタ。
 * doc JSON を value/onChange で制御する。書式は StarterKit の基本ノード/マーク＋画像に対応。
 * スラッシュコマンド・構文ハイライトは後続 PR で足す。
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
  className = '',
}: RichTextEditorProps) {
  // onChange は props で差し替わり得るので ref 越しに最新を呼ぶ（onUpdate クロージャの陳腐化を防ぐ）。
  const onChangeRef = useRef(onChange);
  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  // 画像アップロード関数・editor 参照を ref で持ち、editorProps（paste/drop）や
  // ファイル選択ハンドラから最新の値を参照する（useEditor 初期化時の陳腐化を防ぐ）。
  const onImageUploadRef = useRef(onImageUpload);
  useEffect(() => {
    onImageUploadRef.current = onImageUpload;
  }, [onImageUpload]);
  const editorRef = useRef<Editor | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  // アンマウント（別ノートへ切替）後にアップロードが完了しても挿入しないための番人。
  const mountedRef = useRef(true);
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  // 画像ファイルをアップロードして URL を挿入する。失敗は握りつぶす（通知は呼び出し側）。
  // alt にはファイル名を既定で入れる（代替テキストの UI は後続）。
  const insertImageFile = useCallback(async (file: File) => {
    const upload = onImageUploadRef.current;
    const currentEditor = editorRef.current;
    if (!upload || !currentEditor) return;
    try {
      const url = await upload(file);
      // 別ノートへ切り替えてアンマウント/破棄済みなら、別文書へ誤挿入しない。
      if (!mountedRef.current || currentEditor.isDestroyed) return;
      currentEditor.chain().focus().setImage({ src: url, alt: file.name }).run();
    } catch {
      /* 失敗時の通知は呼び出し側の onImageUpload 内で行う方針。ここでは無視。 */
    }
  }, []);

  // dataTransfer / clipboard から画像ファイルだけ取り出して、選択順どおりに 1 つずつ挿入する。
  const uploadImageFiles = useCallback(
    (files: FileList | null | undefined): boolean => {
      if (!onImageUploadRef.current) return false;
      const images = Array.from(files ?? []).filter((file) => isAcceptedImageMimeType(file.type));
      if (images.length === 0) return false;
      // 並列にすると URL 取得の早い順に挿入され表示順が乱れるため、順次に await して選択順を保つ。
      void (async () => {
        for (const file of images) {
          await insertImageFile(file);
        }
      })();
      return true;
    },
    [insertImageFile],
  );

  // 直近で「これが現在値」とみなしている doc の文字列表現。内容ベースで重複 emit を弾く。
  // tiptap はマウント時に一度 onUpdate を発火する（内容は初期 value と同じ）ため、初期値で初期化して
  // その空振り emit を握りつぶす（読み込み直後に「未保存」へ落ちないようにする）。
  const lastValueRef = useRef(JSON.stringify(value));

  const editor = useEditor({
    editable,
    extensions: [
      StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
      Placeholder.configure({ placeholder }),
      Image.configure({ inline: false, allowBase64: false }),
    ],
    content: value,
    editorProps: {
      attributes: {
        class: 'focus:outline-none',
        role: 'textbox',
        'aria-multiline': 'true',
        'aria-label': ariaLabel,
      },
      // クリップボード/ドロップに画像ファイルがあればアップロードして挿入する。
      handlePaste: (_view, event) => uploadImageFiles(event.clipboardData?.files),
      handleDrop: (_view, event) => {
        if (event.dataTransfer?.files && uploadImageFiles(event.dataTransfer.files)) {
          event.preventDefault();
          return true;
        }
        return false;
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

  const handlePickImage = onImageUpload ? () => fileInputRef.current?.click() : undefined;

  return (
    <div
      className={`overflow-hidden rounded-lg border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] ${className}`}
    >
      {editable && editor && <RichTextEditorToolbar editor={editor} onInsertImage={handlePickImage} />}
      <div className="rte-content prose prose-sm max-w-none px-4">
        <EditorContent editor={editor} />
      </div>
      {onImageUpload && (
        // ツールバーの画像ボタンから開く隠しファイル入力。
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          aria-hidden="true"
          tabIndex={-1}
          onChange={(e) => {
            uploadImageFiles(e.target.files);
            e.target.value = '';
          }}
        />
      )}
      {saveStatus && saveStatus !== 'idle' && (
        <div className="flex justify-end border-t border-[var(--color-surface-3)] px-4 py-1.5">
          <SaveStatusIndicator status={saveStatus} />
        </div>
      )}
    </div>
  );
}
