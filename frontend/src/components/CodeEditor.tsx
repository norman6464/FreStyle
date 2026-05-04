import { useEffect, useRef } from 'react';
import * as monaco from 'monaco-editor';
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import { useTheme } from '../hooks/useTheme';

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
  height?: string;
  readOnly?: boolean;
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
}: CodeEditorProps) {
  const { theme } = useTheme();
  const containerRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    if (!containerRef.current) return;
    const editor = monaco.editor.create(containerRef.current, {
      value: toSafeString(value),
      language,
      theme: theme === 'dark' ? 'vs-dark' : 'vs',
      fontSize: 14,
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      wordWrap: 'on',
      readOnly,
      automaticLayout: true,
      tabSize: 4,
      renderLineHighlight: 'line',
    });
    editorRef.current = editor;
    const sub = editor.onDidChangeModelContent(() => {
      onChangeRef.current(editor.getValue());
    });
    return () => {
      sub.dispose();
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
    monaco.editor.setTheme(theme === 'dark' ? 'vs-dark' : 'vs');
  }, [theme]);

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

  return <div ref={containerRef} style={{ height, width: '100%' }} />;
}
