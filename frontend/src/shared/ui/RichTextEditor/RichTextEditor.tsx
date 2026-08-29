import { useCallback, useEffect, useMemo, useRef } from 'react';
import { EditorContent, useEditor, type Editor } from '@tiptap/react';
import { ACCEPTED_IMAGE_ACCEPT_ATTR } from '@/shared/config/imageUpload';
import { createEditorExtensions } from './editorExtensions';
import type { EditorCommand } from './editorCommands';
import { buildSlashItems } from './slashItems';
import { acceptedImageFiles, insertUploadedImages } from './imageInsertion';
import { sanitizeDocLinks } from './linkSafety';
import { openClickedLink } from './linkClick';
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
  /**
   * '/' メニューへ追加するコマンド。エディタが知らない操作（子ページの作成など、
   * 業務を知る操作）は呼び出し側がここから差し込む。エディタは項目の中身を解釈しない。
   *
   * **項目はエディタ生成時に固定される。** run の closure が画面の状態を読むなら、
   * 呼び出し側で ref 越しに最新を参照させること（この配列自体を差し替えても反映されない）。
   */
  extraSlashCommands?: EditorCommand[];
  /**
   * 書式ボタン列を上部に常設する（編集できるときだけ）。バブルメニュー（選択時に
   * 浮かぶ方）はそのまま併存する — どちらも同じコマンドレジストリを叩くので二重実装にはならない。
   */
  /**
   * 本文中の内部ページリンク（/p/{id}）を開くときの遷移。渡すとアプリ内遷移になる
   * （渡さなければ素の遷移）。外部リンクは常に新しいタブで開く。
   */
  onNavigateToPage?: (path: string) => void;
  /** 増えたら本文の先頭へフォーカスを移す合図（題名で Enter → 本文へ、のため）。 */
  focusSignal?: number;
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
/**
 * stableDocString は doc の同一性比較のためにキー順を揃えて文字列化する。
 *
 * 素の JSON.stringify で比べると、**キーの並びが違うだけ**で「内容が変わった」と
 * 誤判定する。実際、サーバーはページ参照の題名解決で doc を作り直して返し、その際
 * キーがアルファベット順になる。tiptap の getJSON() は type が先なので、素の比較だと
 * 開いただけで onChange が発火し、閲覧しただけの人が本文の保存（全置換）を発行して
 * 同時編集者の直近の書き込みを潰しうる。比較のためだけに使い、値そのものは変えない。
 */
function stableDocString(node: unknown): string {
  if (Array.isArray(node)) {
    return `[${node.map(stableDocString).join(',')}]`;
  }
  if (node !== null && typeof node === 'object') {
    const entries = Object.entries(node as Record<string, unknown>)
      .filter(([, v]) => v !== undefined)
      .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0))
      .map(([k, v]) => `${JSON.stringify(k)}:${stableDocString(v)}`);
    return `{${entries.join(',')}}`;
  }
  return JSON.stringify(node) ?? 'null';
}

export default function RichTextEditor({
  value,
  onChange,
  editable = true,
  placeholder = '本文を入力…',
  ariaLabel = '本文',
  saveStatus,
  onImageUpload,
  onCreate,
  extraSlashCommands,
  onNavigateToPage,
  focusSignal = 0,
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
  // '/image' から開くファイル選択（キーボード/クリックでも画像を挿入できる経路）。
  const fileInputRef = useRef<HTMLInputElement>(null);

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
  const lastValueRef = useRef(stableDocString(value));

  // '/' メニューの項目。ベースはレジストリ（ブロック変換＋挿入）。画像アップロードが
  // 配線されているときだけ /image（ファイル選択）を足す。onImageUpload の有無だけに依存させ、
  // 拡張一式が編集のたびに作り直されないようにする。
  const hasImageUpload = Boolean(onImageUpload);
  const slashItems = useMemo<EditorCommand[]>(() => {
    const extra: EditorCommand[] = hasImageUpload
      ? [
          {
            id: 'image',
            label: '画像',
            group: 'insert',
            glyph: '🖼',
            keywords: ['image', 'img', 'photo', 'picture', 'upload'],
            run: () => fileInputRef.current?.click(),
          },
        ]
      : [];
    return buildSlashItems([...extra, ...(extraSlashCommands ?? [])]);
    // extraSlashCommands は「エディタ生成時に固定」の契約（props の JSDoc 参照）なので
    // 依存に入れない — 入れても extensions は作り直されず、揃わない再計算だけが増える。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hasImageUpload]);

  const editor = useEditor({
    editable,
    extensions: createEditorExtensions({ placeholder, slashItems }),
    // 読み込み側のリンク洗浄。doc JSON は API から丸ごと差し込めるので、エディタの入力・貼り付けを
    // どれだけ固めても「危険な href がすでに入った doc」はここから入ってくる。開いた時点で落とす。
    content: sanitizeDocLinks(value),
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
      // 保存側のリンク洗浄。表示のときだけ無害化する作りだと、DB には危険な href が残ったままになり、
      // 別の読み手（別のクライアント・API 直叩き）に対して無防備なままになる。外へ出す値を洗う。
      const next = sanitizeDocLinks(currentEditor.getJSON()) as RichDocContent;
      const nextStr = stableDocString(next);
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
    const valueStr = stableDocString(value);
    if (valueStr !== lastValueRef.current) {
      lastValueRef.current = valueStr;
      // 差し替えで入ってくる doc も読み込み時と同じ経路で洗う（別ドキュメントへの切り替え）。
      editor.commands.setContent(sanitizeDocLinks(value), { emitUpdate: false });
    }
  }, [editor, value]);

  // editable の変更を反映する。
  useEffect(() => {
    editor?.setEditable(editable);
  }, [editor, editable]);

  // 「増えたときだけ」フォーカスを移す。マウント時の値では動かない — ページを
  // 開き直しただけで本文が奪ってしまわないため（サイドバーの openSignal と同じ形）。
  const seenFocusSignal = useRef(focusSignal);
  useEffect(() => {
    // editor がまだ無いときは**合図を消費しない**。ここで見たことにすると、
    // 初期化中に題名で Enter を押した合図が捨てられ、本文へ移らないまま終わる。
    if (!editor) return;
    if (focusSignal > seenFocusSignal.current) {
      editor.commands.focus('start');
    }
    seenFocusSignal.current = focusSignal;
  }, [editor, focusSignal]);

  return (
    // リンクはクリックで開く（編集中も読み取り専用も同じ経路。linkClick.ts のコメント参照）。
    // preventDefault は読み取り専用の素の <a> の既定遷移（全画面リロード）を止めるため。
    // キーボードは別経路が既にある: <a> 上の Enter はブラウザが click として発火する。
    <div
      className={`rte-root ${className}`}
      onClick={(event) => {
        if (openClickedLink(event.nativeEvent, onNavigateToPage, { editable })) {
          event.preventDefault();
        }
      }}
    >
      <div className="rte-content prose max-w-none">
        <EditorContent editor={editor} />
      </div>
      {editable && editor && <BubbleFormatMenu editor={editor} />}
      {onImageUpload && (
        // '/image' から開く隠しファイル入力（DnD/貼り付けと同じ挿入経路へ流す）。
        <input
          ref={fileInputRef}
          type="file"
          accept={ACCEPTED_IMAGE_ACCEPT_ATTR}
          className="hidden"
          aria-hidden="true"
          tabIndex={-1}
          onChange={(e) => {
            handleImageFiles(e.target.files);
            e.target.value = '';
          }}
        />
      )}
      {saveStatus && saveStatus !== 'idle' && (
        <div className="mt-2 flex justify-end">
          <SaveStatusIndicator status={saveStatus} />
        </div>
      )}
    </div>
  );
}
