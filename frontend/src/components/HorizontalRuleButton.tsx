import { MinusIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface HorizontalRuleButtonProps {
  onHorizontalRule: () => void;
}

export default function HorizontalRuleButton({ onHorizontalRule }: HorizontalRuleButtonProps) {
  return <ToolbarIconButton icon={MinusIcon} label="水平線" onClick={onHorizontalRule} />;
}
