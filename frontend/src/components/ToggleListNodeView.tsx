import { NodeViewWrapper, NodeViewContent } from '@tiptap/react';
import { ChevronRightIcon } from '@heroicons/react/24/solid';

export default function ToggleListNodeView({ node, updateAttributes }: {
  node: { attrs: { open: boolean } };
  updateAttributes: (attrs: { open: boolean }) => void;
}) {
  const isOpen = node.attrs.open;

  return (
    <NodeViewWrapper
      className={`toggle-list my-1 ${isOpen ? 'toggle-open' : 'toggle-closed'}`}
      data-toggle-open={isOpen}
    >
      <div className="flex items-start">
        <button
          type="button"
          contentEditable={false}
          aria-label={isOpen ? 'トグルを閉じる' : 'トグルを開く'}
          aria-expanded={isOpen}
          className="mt-0.5 w-5 h-5 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] hover:text-[var(--color-text-secondary)] shrink-0"
          onClick={() => updateAttributes({ open: !isOpen })}
        >
          <ChevronRightIcon
            className={`w-3.5 h-3.5 transition-transform duration-150 ${isOpen ? 'rotate-90' : ''}`}
          />
        </button>
        <div className="flex-1 min-w-0">
          <NodeViewContent />
        </div>
      </div>
    </NodeViewWrapper>
  );
}
