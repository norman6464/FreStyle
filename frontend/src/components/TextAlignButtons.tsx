import { Bars3BottomLeftIcon, Bars3Icon, Bars3BottomRightIcon } from '@heroicons/react/24/solid';

type Alignment = 'left' | 'center' | 'right';

interface TextAlignButtonsProps {
  onAlign: (alignment: Alignment) => void;
  activeAlign?: Alignment;
}

const ALIGNMENTS: { alignment: Alignment; label: string; Icon: typeof Bars3Icon }[] = [
  { alignment: 'left', label: '左寄せ', Icon: Bars3BottomLeftIcon },
  { alignment: 'center', label: '中央寄せ', Icon: Bars3Icon },
  { alignment: 'right', label: '右寄せ', Icon: Bars3BottomRightIcon },
];

export default function TextAlignButtons({ onAlign, activeAlign }: TextAlignButtonsProps) {
  return (
    <div className="flex items-center gap-0.5">
      {ALIGNMENTS.map(({ alignment, label, Icon }) => (
        <button
          key={alignment}
          type="button"
          aria-label={label}
          className={`w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] transition-colors ${
            activeAlign === alignment ? 'text-primary-500' : 'text-[var(--color-text-faint)]'
          }`}
          onClick={() => onAlign(alignment)}
        >
          <Icon className="w-4 h-4" />
        </button>
      ))}
    </div>
  );
}
