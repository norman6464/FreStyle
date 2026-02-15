import type { EditorFormatHandlers } from '../hooks/useEditorFormat';
import FormatButtons from './FormatButtons';
import ColorPicker from './ColorPicker';
import HighlightPicker from './HighlightPicker';
import TextAlignButtons from './TextAlignButtons';
import UndoRedoButtons from './UndoRedoButtons';
import IndentButtons from './IndentButtons';
import BlockquoteButton from './BlockquoteButton';
import HorizontalRuleButton from './HorizontalRuleButton';
import CodeBlockButton from './CodeBlockButton';
import InlineCodeButton from './InlineCodeButton';
import ListButtons from './ListButtons';
import ClearFormattingButton from './ClearFormattingButton';
import KeyboardShortcutsHelp from './KeyboardShortcutsHelp';

interface EditorToolbarProps {
  handlers: EditorFormatHandlers;
}

export default function EditorToolbar({ handlers }: EditorToolbarProps) {
  return (
    <div className="px-8 py-1 border-b border-[var(--color-surface-3)] flex items-center gap-3">
      <FormatButtons onBold={handlers.handleBold} onItalic={handlers.handleItalic} onUnderline={handlers.handleUnderline} onStrike={handlers.handleStrike} onSuperscript={handlers.handleSuperscript} onSubscript={handlers.handleSubscript} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <ColorPicker onSelectColor={handlers.handleSelectColor} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <HighlightPicker onSelectHighlight={handlers.handleHighlight} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <TextAlignButtons onAlign={handlers.handleAlign} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <IndentButtons onIndent={handlers.handleIndent} onOutdent={handlers.handleOutdent} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <BlockquoteButton onBlockquote={handlers.handleBlockquote} />
      <HorizontalRuleButton onHorizontalRule={handlers.handleHorizontalRule} />
      <InlineCodeButton onInlineCode={handlers.handleInlineCode} />
      <CodeBlockButton onCodeBlock={handlers.handleCodeBlock} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <ListButtons onBulletList={handlers.handleBulletList} onOrderedList={handlers.handleOrderedList} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <UndoRedoButtons onUndo={handlers.handleUndo} onRedo={handlers.handleRedo} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <ClearFormattingButton onClearFormatting={handlers.handleClearFormatting} />
      <div className="w-px h-4 bg-[var(--color-surface-3)]" />
      <KeyboardShortcutsHelp />
    </div>
  );
}
