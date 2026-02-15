import { useState } from 'react';
import { NodeViewWrapper, NodeViewContent } from '@tiptap/react';
import { ClipboardIcon, CheckIcon } from '@heroicons/react/24/outline';

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

  const handleCopy = () => {
    navigator.clipboard.writeText(node.textContent);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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
      <pre>
        <NodeViewContent as="code" />
      </pre>
    </NodeViewWrapper>
  );
}
