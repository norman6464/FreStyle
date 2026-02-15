import { BoldIcon, ItalicIcon, UnderlineIcon, StrikethroughIcon } from '@heroicons/react/24/solid';
import { ArrowUpIcon, ArrowDownIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface FormatButtonsProps {
  onBold: () => void;
  onItalic: () => void;
  onUnderline: () => void;
  onStrike: () => void;
  onSuperscript: () => void;
  onSubscript: () => void;
}

const BUTTONS = [
  { label: '太字', Icon: BoldIcon, key: 'bold' as const },
  { label: '斜体', Icon: ItalicIcon, key: 'italic' as const },
  { label: '下線', Icon: UnderlineIcon, key: 'underline' as const },
  { label: '取り消し線', Icon: StrikethroughIcon, key: 'strike' as const },
  { label: '上付き文字', Icon: ArrowUpIcon, key: 'superscript' as const },
  { label: '下付き文字', Icon: ArrowDownIcon, key: 'subscript' as const },
];

export default function FormatButtons({ onBold, onItalic, onUnderline, onStrike, onSuperscript, onSubscript }: FormatButtonsProps) {
  const handlers = { bold: onBold, italic: onItalic, underline: onUnderline, strike: onStrike, superscript: onSuperscript, subscript: onSubscript };

  return (
    <div className="flex items-center gap-0.5">
      {BUTTONS.map(({ label, Icon, key }) => (
        <ToolbarIconButton key={key} icon={Icon} label={label} onClick={handlers[key]} />
      ))}
    </div>
  );
}
