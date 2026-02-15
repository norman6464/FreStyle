import { BoldIcon, ItalicIcon, UnderlineIcon, StrikethroughIcon } from '@heroicons/react/24/solid';

interface FormatButtonsProps {
  onBold: () => void;
  onItalic: () => void;
  onUnderline: () => void;
  onStrike: () => void;
}

const BUTTONS = [
  { label: '太字', Icon: BoldIcon, key: 'bold' as const },
  { label: '斜体', Icon: ItalicIcon, key: 'italic' as const },
  { label: '下線', Icon: UnderlineIcon, key: 'underline' as const },
  { label: '取り消し線', Icon: StrikethroughIcon, key: 'strike' as const },
];

export default function FormatButtons({ onBold, onItalic, onUnderline, onStrike }: FormatButtonsProps) {
  const handlers = { bold: onBold, italic: onItalic, underline: onUnderline, strike: onStrike };

  return (
    <div className="flex items-center gap-0.5">
      {BUTTONS.map(({ label, Icon, key }) => (
        <button
          key={key}
          type="button"
          aria-label={label}
          className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
          onClick={handlers[key]}
        >
          <Icon className="w-4 h-4" />
        </button>
      ))}
    </div>
  );
}
