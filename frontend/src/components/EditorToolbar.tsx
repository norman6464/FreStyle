import ColorPicker from './ColorPicker';
import TextAlignButtons from './TextAlignButtons';

interface EditorToolbarProps {
  onSelectColor: (color: string) => void;
  onAlign: (alignment: 'left' | 'center' | 'right') => void;
}

export default function EditorToolbar({ onSelectColor, onAlign }: EditorToolbarProps) {
  return (
    <div className="px-8 py-1 border-b border-[var(--color-surface-3)] flex items-center gap-3">
      <ColorPicker onSelectColor={onSelectColor} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <TextAlignButtons onAlign={onAlign} />
    </div>
  );
}
