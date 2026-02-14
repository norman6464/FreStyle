import { useState } from 'react';

interface Message {
  id: number;
  content: string;
  role: 'user' | 'assistant';
}

interface ExportSessionButtonProps {
  messages: Message[];
}

export default function ExportSessionButton({ messages }: ExportSessionButtonProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const text = messages
      .map((msg) => `${msg.role === 'user' ? 'あなた' : 'AI'}: ${msg.content}`)
      .join('\n\n');

    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={handleCopy}
      disabled={messages.length === 0}
      title={copied ? 'コピーしました' : '会話をコピー'}
      className="p-1.5 rounded-lg hover:bg-surface-2 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    >
      {copied ? (
        <svg className="w-4 h-4 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      ) : (
        <svg className="w-4 h-4 text-[var(--color-text-muted)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
      )}
    </button>
  );
}
