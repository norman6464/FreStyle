import { XMarkIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface ClearFormattingButtonProps {
  onClearFormatting: () => void;
}

export default function ClearFormattingButton({ onClearFormatting }: ClearFormattingButtonProps) {
  return <ToolbarIconButton icon={XMarkIcon} label="書式クリア" onClick={onClearFormatting} />;
}
