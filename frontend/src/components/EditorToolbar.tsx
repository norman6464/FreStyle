import type { EditorFormatHandlers } from '../hooks/useEditorFormat';
import HeadingSelect from './HeadingSelect';
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
import TaskListButton from './TaskListButton';
import ClearFormattingButton from './ClearFormattingButton';
import KeyboardShortcutsHelp from './KeyboardShortcutsHelp';
import ToolbarDivider from './ToolbarDivider';

interface EditorToolbarProps {
  handlers: EditorFormatHandlers;
}

export default function EditorToolbar({ handlers }: EditorToolbarProps) {
  return (
    <div className="px-8 py-1 border-b border-[var(--color-surface-3)] flex items-center gap-3">
      <HeadingSelect onHeading={handlers.handleHeading} />
      <ToolbarDivider />
      <FormatButtons onBold={handlers.handleBold} onItalic={handlers.handleItalic} onUnderline={handlers.handleUnderline} onStrike={handlers.handleStrike} onSuperscript={handlers.handleSuperscript} onSubscript={handlers.handleSubscript} />
      <ToolbarDivider />
      <ColorPicker onSelectColor={handlers.handleSelectColor} />
      <ToolbarDivider />
      <HighlightPicker onSelectHighlight={handlers.handleHighlight} />
      <ToolbarDivider />
      <TextAlignButtons onAlign={handlers.handleAlign} />
      <ToolbarDivider />
      <IndentButtons onIndent={handlers.handleIndent} onOutdent={handlers.handleOutdent} />
      <ToolbarDivider />
      <BlockquoteButton onBlockquote={handlers.handleBlockquote} />
      <HorizontalRuleButton onHorizontalRule={handlers.handleHorizontalRule} />
      <InlineCodeButton onInlineCode={handlers.handleInlineCode} />
      <CodeBlockButton onCodeBlock={handlers.handleCodeBlock} />
      <ToolbarDivider />
      <ListButtons onBulletList={handlers.handleBulletList} onOrderedList={handlers.handleOrderedList} />
      <TaskListButton onTaskList={handlers.handleTaskList} />
      <ToolbarDivider />
      <UndoRedoButtons onUndo={handlers.handleUndo} onRedo={handlers.handleRedo} />
      <ToolbarDivider />
      <ClearFormattingButton onClearFormatting={handlers.handleClearFormatting} />
      <ToolbarDivider />
      <KeyboardShortcutsHelp />
    </div>
  );
}
