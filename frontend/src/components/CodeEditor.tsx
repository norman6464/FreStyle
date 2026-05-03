import { Editor } from '@monaco-editor/react';
import { useTheme } from '../hooks/useTheme';

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string;
  height?: string;
  readOnly?: boolean;
}

/**
 * CodeEditor — Monaco Editor (VS Code エンジン) のラッパーコンポーネント。
 * テーマは useTheme に追従し、dark/light を自動切り替えする。
 */
export default function CodeEditor({
  value,
  onChange,
  language = 'php',
  height = '400px',
  readOnly = false,
}: CodeEditorProps) {
  const { theme } = useTheme();

  return (
    <Editor
      height={height}
      language={language}
      value={value}
      theme={theme === 'dark' ? 'vs-dark' : 'vs'}
      onChange={(v) => onChange(v ?? '')}
      options={{
        fontSize: 14,
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        wordWrap: 'on',
        readOnly,
        automaticLayout: true,
        tabSize: 4,
        renderLineHighlight: 'line',
      }}
    />
  );
}
