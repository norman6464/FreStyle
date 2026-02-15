import { ArrowUturnLeftIcon, ArrowUturnRightIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface UndoRedoButtonsProps {
  onUndo: () => void;
  onRedo: () => void;
}

export default function UndoRedoButtons({ onUndo, onRedo }: UndoRedoButtonsProps) {
  return (
    <div className="flex items-center gap-0.5">
      <ToolbarIconButton icon={ArrowUturnLeftIcon} label="元に戻す" onClick={onUndo} />
      <ToolbarIconButton icon={ArrowUturnRightIcon} label="やり直す" onClick={onRedo} />
    </div>
  );
}
