import { useState } from 'react';
import { CheckIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline';
import { UI_TIMINGS } from '../constants/uiTimings';

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

    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), UI_TIMINGS.COPY_FEEDBACK_DURATION);
    } catch {
      // Clipboard API非対応・権限なし時は無視
    }
  };

  return (
    <button
      onClick={handleCopy}
      disabled={messages.length === 0}
      title={copied ? 'コピーしました' : '会話をコピー'}
      aria-label={copied ? 'コピーしました' : '会話をコピー'}
      className="p-1.5 rounded-lg hover:bg-surface-2 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    >
      {copied ? (
        <CheckIcon className="w-4 h-4 text-emerald-500" />
      ) : (
        <ClipboardDocumentIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
      )}
    </button>
  );
}
