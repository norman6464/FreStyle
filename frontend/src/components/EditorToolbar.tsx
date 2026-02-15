import FormatButtons from './FormatButtons';
import ColorPicker from './ColorPicker';
import HighlightPicker from './HighlightPicker';
import TextAlignButtons from './TextAlignButtons';
import UndoRedoButtons from './UndoRedoButtons';

interface EditorToolbarProps {
  onBold: () => void;
  onItalic: () => void;
  onUnderline: () => void;
  onStrike: () => void;
  onSuperscript: () => void;
  onSubscript: () => void;
  onSelectColor: (color: string) => void;
  onHighlight: (color: string) => void;
  onAlign: (alignment: 'left' | 'center' | 'right') => void;
  onUndo: () => void;
  onRedo: () => void;
}

export default function EditorToolbar({
  onBold,
  onItalic,
  onUnderline,
  onStrike,
  onSuperscript,
  onSubscript,
  onSelectColor,
  onHighlight,
  onAlign,
  onUndo,
  onRedo,
}: EditorToolbarProps) {
  return (
    <div className="px-8 py-1 border-b border-[var(--color-surface-3)] flex items-center gap-3">
      <FormatButtons onBold={onBold} onItalic={onItalic} onUnderline={onUnderline} onStrike={onStrike} onSuperscript={onSuperscript} onSubscript={onSubscript} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <ColorPicker onSelectColor={onSelectColor} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <HighlightPicker onSelectHighlight={onHighlight} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <TextAlignButtons onAlign={onAlign} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <UndoRedoButtons onUndo={onUndo} onRedo={onRedo} />
    </div>
  );
}
