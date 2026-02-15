import { ListBulletIcon, NumberedListIcon } from '@heroicons/react/24/outline';
import ToolbarIconButton from './ToolbarIconButton';

interface ListButtonsProps {
  onBulletList: () => void;
  onOrderedList: () => void;
}

export default function ListButtons({ onBulletList, onOrderedList }: ListButtonsProps) {
  return (
    <>
      <ToolbarIconButton icon={ListBulletIcon} label="箇条書き" onClick={onBulletList} />
      <ToolbarIconButton icon={NumberedListIcon} label="番号付きリスト" onClick={onOrderedList} />
    </>
  );
}
