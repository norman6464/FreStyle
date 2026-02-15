import { Bars3BottomLeftIcon, Bars3Icon, Bars3BottomRightIcon } from '@heroicons/react/24/solid';
import ToolbarIconButton from './ToolbarIconButton';

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
        <ToolbarIconButton
          key={alignment}
          icon={Icon}
          label={label}
          onClick={() => onAlign(alignment)}
          isActive={activeAlign === alignment}
        />
      ))}
    </div>
  );
}
