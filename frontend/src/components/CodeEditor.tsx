import { useEffect, useRef, useState } from 'react';
import * as monaco from 'monaco-editor';
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';

// CDN 読み込みを完全回避: Vite バンドル済み monaco-editor を直接使う。
// @monaco-editor/react は loader が jsDelivr を参照するため使わない。
if (typeof self !== 'undefined' && !self.MonacoEnvironment) {
  self.MonacoEnvironment = {
    getWorker() {
      return new editorWorker();
    },
  };
}

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string;
  /** autoGrow=true (default) のときは ignored、 fixed height にしたい場合だけ指定 */
  height?: string;
  readOnly?: boolean;
  /**
   * autoGrow が true (default) のとき、 エディタは内部スクロールせず コンテンツの行数
   * に合わせて 縦に伸びる。 ページ側でスクロールする UX (paiza / Wandbox 風)。
   */
  autoGrow?: boolean;
  /** autoGrow 時の最小高さ (px)。 空のエディタ時の見た目用。 */
  minHeight?: number;
  /** autoGrow 時の最大高さ (px)。 巨大入力で 画面を埋め尽くさないための上限。 */
  maxHeight?: number;
  /** Ctrl+Enter / Cmd+Enter で呼ばれる実行ハンドラ（コード実行ショートカット）。 */
  onRun?: () => void;
}

// Monaco の setValue / create({value}) は引数が string でない場合 "Illegal argument" を throw する。
// API レスポンスが null や undefined を返す可能性があるため必ず string に正規化する。
function toSafeString(v: unknown): string {
  if (typeof v === 'string') return v;
  if (v == null) return '';
  return String(v);
}

export default function CodeEditor({
  value,
  onChange,
  language = 'php',
  height = '400px',
  readOnly = false,
  autoGrow = true,
  minHeight = 240,
  maxHeight = 2000,
  onRun,
}: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  // 最新の onRun を ref で保持し、エディタ再生成なしにショートカットから呼べるようにする。
  const onRunRef = useRef(onRun);
  onRunRef.current = onRun;

  // autoGrow=true のときの 動的 height (px)。
  const [autoHeight, setAutoHeight] = useState<number>(minHeight);

  useEffect(() => {
    if (!containerRef.current) return;
    const editor = monaco.editor.create(containerRef.current, {
      value: toSafeString(value),
      language,
      // コードエディタは常にダークテーマで統一（ライト/ダークどちらの UI でも視認性確保）
      theme: 'vs-dark',
      fontSize: 14,
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      wordWrap: 'on',
      readOnly,
      automaticLayout: true,
      tabSize: 4,
      renderLineHighlight: 'line',
      // Monaco は default で mouse wheel イベントを 常に消費するため、 エディタが
      // 短くて 内部スクロールが要らない状態でも、 ページ全体のスクロールが効かない
      // 不便な挙動を起こす。 alwaysConsumeMouseWheel=false にすると エディタ内に
      // スクロール余地が無いとき wheel イベントを 親に bubble させて ページが普通に
      // スクロールできるようにする。
      scrollbar: {
        alwaysConsumeMouseWheel: false,
      },
    });
    editorRef.current = editor;

    // Ctrl+Enter / Cmd+Enter でコードを実行する（CtrlCmd は Win/Linux=Ctrl, Mac=Cmd）。
    // addCommand は ESM バンドルの monaco だとキーバインドが発火しないことがあるため、
    // 公式に推奨される addAction を使う（キーバインド + 右クリックメニュー + コマンドパレットに登録）。
    editor.addAction({
      id: 'frestyle.runCode',
      label: 'コードを実行 (Ctrl/Cmd + Enter)',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
      contextMenuGroupId: 'navigation',
      contextMenuOrder: 1,
      run: () => {
        onRunRef.current?.();
      },
    });

    const sub = editor.onDidChangeModelContent(() => {
      onChangeRef.current(editor.getValue());
    });

    // autoGrow: コンテンツのサイズ変更を監視して container height を更新する。
    // Monaco の getContentHeight() は 「全行表示に必要な高さ (px)」 を返す。
    // これに余白 (~16px) を足して 最小 / 最大で clamp する。
    let sizeSub: monaco.IDisposable | null = null;
    if (autoGrow) {
      const updateHeight = () => {
        const contentHeight = editor.getContentHeight();
        const next = Math.min(maxHeight, Math.max(minHeight, contentHeight + 16));
        setAutoHeight(next);
        // editor 内部レイアウトを 更新後の高さに合わせる
        editor.layout({
          width: containerRef.current?.clientWidth ?? 0,
          height: next,
        });
      };
      updateHeight();
      sizeSub = editor.onDidContentSizeChange(updateHeight);
    }

    return () => {
      sub.dispose();
      sizeSub?.dispose();
      editor.dispose();
      editorRef.current = null;
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const model = editorRef.current?.getModel();
    if (model) monaco.editor.setModelLanguage(model, language);
  }, [language]);

  useEffect(() => {
    editorRef.current?.updateOptions({ readOnly });
  }, [readOnly]);

  useEffect(() => {
    const editor = editorRef.current;
    if (!editor) return;
    // model が dispose 済みだと setValue が "Illegal argument" を投げるため事前チェック。
    const model = editor.getModel();
    if (!model || model.isDisposed()) return;
    const safeValue = toSafeString(value);
    if (editor.getValue() !== safeValue) {
      editor.setValue(safeValue);
    }
  }, [value]);

  // autoGrow のときは動的 height、 そうでなければ props の height を使う。
  const containerHeight = autoGrow ? `${autoHeight}px` : height;

  return <div ref={containerRef} style={{ height: containerHeight, width: '100%' }} />;
}
