import { ChatBubbleBottomCenterTextIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface BlockquoteButtonProps {
  onBlockquote: () => void;
}

export default function BlockquoteButton({ onBlockquote }: BlockquoteButtonProps) {
  return <ToolbarIconButton icon={ChatBubbleBottomCenterTextIcon} label="引用" onClick={onBlockquote} />;
}
