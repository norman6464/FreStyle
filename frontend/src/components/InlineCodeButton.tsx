import { CommandLineIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface InlineCodeButtonProps {
  onInlineCode: () => void;
}

export default function InlineCodeButton({ onInlineCode }: InlineCodeButtonProps) {
  return <ToolbarIconButton icon={CommandLineIcon} label="インラインコード" onClick={onInlineCode} />;
}
