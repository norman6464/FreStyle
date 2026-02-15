import { TableCellsIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface TableInsertButtonProps {
  onInsertTable: () => void;
}

export default function TableInsertButton({ onInsertTable }: TableInsertButtonProps) {
  return <ToolbarIconButton icon={TableCellsIcon} label="テーブル挿入" onClick={onInsertTable} />;
}
