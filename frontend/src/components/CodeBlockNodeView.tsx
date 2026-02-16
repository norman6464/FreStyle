import { useState, useMemo } from 'react';
import { NodeViewWrapper, NodeViewContent } from '@tiptap/react';
import { ClipboardIcon, CheckIcon } from '@heroicons/react/24/outline';
import { UI_TIMINGS } from '../constants/uiTimings';

const LANGUAGES = [
  { value: '', label: 'プレーンテキスト' },
  { value: 'javascript', label: 'JavaScript' },
  { value: 'typescript', label: 'TypeScript' },
  { value: 'python', label: 'Python' },
  { value: 'java', label: 'Java' },
  { value: 'css', label: 'CSS' },
  { value: 'html', label: 'HTML' },
  { value: 'json', label: 'JSON' },
  { value: 'bash', label: 'Bash' },
  { value: 'sql', label: 'SQL' },
  { value: 'xml', label: 'XML' },
  { value: 'markdown', label: 'Markdown' },
  { value: 'yaml', label: 'YAML' },
  { value: 'rust', label: 'Rust' },
  { value: 'go', label: 'Go' },
  { value: 'c', label: 'C' },
  { value: 'cpp', label: 'C++' },
  { value: 'csharp', label: 'C#' },
  { value: 'php', label: 'PHP' },
  { value: 'ruby', label: 'Ruby' },
  { value: 'swift', label: 'Swift' },
  { value: 'kotlin', label: 'Kotlin' },
];

interface CodeBlockNodeViewProps {
  node: { attrs: { language: string | null }; textContent: string };
  updateAttributes: (attrs: { language: string }) => void;
}

export default function CodeBlockNodeView({ node, updateAttributes }: CodeBlockNodeViewProps) {
  const [copied, setCopied] = useState(false);
  const language = node.attrs.language || '';

  const lineCount = useMemo(() => {
    const text = node.textContent;
    if (!text) return 1;
    return text.split('\n').length;
  }, [node.textContent]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(node.textContent);
      setCopied(true);
      setTimeout(() => setCopied(false), UI_TIMINGS.COPY_FEEDBACK_DURATION);
    } catch {
      // Clipboard API非対応・権限なし時は無視
    }
  };

  return (
    <NodeViewWrapper className="code-block-wrapper">
      <div className="code-block-header" contentEditable={false}>
        <select
          value={language}
          onChange={(e) => updateAttributes({ language: e.target.value })}
          className="code-block-language"
          aria-label="プログラミング言語"
        >
          {LANGUAGES.map(lang => (
            <option key={lang.value} value={lang.value}>{lang.label}</option>
          ))}
        </select>
        <button
          type="button"
          onClick={handleCopy}
          className="code-block-copy"
          aria-label="コードをコピー"
        >
          {copied ? (
            <><CheckIcon className="w-3.5 h-3.5" /><span>コピーしました</span></>
          ) : (
            <><ClipboardIcon className="w-3.5 h-3.5" /><span>コピー</span></>
          )}
        </button>
      </div>
      <div className="code-block-body">
        <div className="code-block-line-numbers" contentEditable={false} aria-hidden="true">
          {Array.from({ length: lineCount }, (_, i) => (
            <span key={i}>{i + 1}</span>
          ))}
        </div>
        <pre>
          <NodeViewContent as="code" />
        </pre>
      </div>
    </NodeViewWrapper>
  );
}
