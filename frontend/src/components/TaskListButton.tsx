import { QueueListIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface TaskListButtonProps {
  onTaskList: () => void;
}

export default function TaskListButton({ onTaskList }: TaskListButtonProps) {
  return <ToolbarIconButton icon={QueueListIcon} label="タスクリスト" onClick={onTaskList} />;
}
