import { ArrowRightIcon, ArrowLeftIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface IndentButtonsProps {
  onIndent: () => void;
  onOutdent: () => void;
}

export default function IndentButtons({ onIndent, onOutdent }: IndentButtonsProps) {
  return (
    <div className="flex items-center gap-0.5">
      <ToolbarIconButton icon={ArrowLeftIcon} label="インデント減少" onClick={onOutdent} />
      <ToolbarIconButton icon={ArrowRightIcon} label="インデント増加" onClick={onIndent} />
    </div>
  );
}
