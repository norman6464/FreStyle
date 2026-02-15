import { CodeBracketIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface CodeBlockButtonProps {
  onCodeBlock: () => void;
}

export default function CodeBlockButton({ onCodeBlock }: CodeBlockButtonProps) {
  return <ToolbarIconButton icon={CodeBracketIcon} label="コードブロック" onClick={onCodeBlock} />;
}
