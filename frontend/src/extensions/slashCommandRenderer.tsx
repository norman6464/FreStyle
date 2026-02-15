import { createRoot, type Root } from 'react-dom/client';
import type { SuggestionProps, SuggestionKeyDownProps } from '@tiptap/suggestion';
import tippy, { type Instance as TippyInstance } from 'tippy.js';
import SlashCommandMenu from '../components/SlashCommandMenu';
import type { SlashCommand } from '../constants/slashCommands';

interface RendererState {
  items: SlashCommand[];
  selectedIndex: number;
  onSelect: (index: number) => void;
}

export function slashCommandRenderer() {
  let root: Root | null = null;
  let popup: TippyInstance | null = null;
  let container: HTMLDivElement | null = null;
  let state: RendererState = { items: [], selectedIndex: 0, onSelect: () => {} };

  function render() {
    if (!root) return;
    root.render(
      <SlashCommandMenu
        items={state.items}
        selectedIndex={state.selectedIndex}
        onSelect={state.onSelect}
      />
    );
  }

  return {
    onStart: (props: SuggestionProps<SlashCommand>) => {
      container = document.createElement('div');
      root = createRoot(container);

      state = {
        items: props.items,
        selectedIndex: 0,
        onSelect: (index: number) => {
          const item = props.items[index];
          if (item) props.command(item);
        },
      };

      render();

      popup = tippy(document.body, {
        getReferenceClientRect: () => props.clientRect?.() ?? new DOMRect(),
        appendTo: () => document.body,
        content: container,
        showOnCreate: true,
        interactive: true,
        trigger: 'manual',
        placement: 'bottom-start',
        maxWidth: 'none',
      });
    },

    onUpdate: (props: SuggestionProps<SlashCommand>) => {
      state = {
        ...state,
        items: props.items,
        selectedIndex: 0,
        onSelect: (index: number) => {
          const item = props.items[index];
          if (item) props.command(item);
        },
      };
      render();

      popup?.setProps({
        getReferenceClientRect: () => props.clientRect?.() ?? new DOMRect(),
      });
    },

    onKeyDown: (props: SuggestionKeyDownProps) => {
      const { event } = props;

      if (event.key === 'ArrowDown') {
        state = {
          ...state,
          selectedIndex: (state.selectedIndex + 1) % state.items.length,
        };
        render();
        return true;
      }

      if (event.key === 'ArrowUp') {
        state = {
          ...state,
          selectedIndex: (state.selectedIndex - 1 + state.items.length) % state.items.length,
        };
        render();
        return true;
      }

      if (event.key === 'Enter') {
        const item = state.items[state.selectedIndex];
        if (item) {
          const commandProps = (props as unknown as { command: (item: SlashCommand) => void });
          commandProps.command(item);
        }
        return true;
      }

      if (event.key === 'Escape') {
        popup?.hide();
        return true;
      }

      return false;
    },

    onExit: () => {
      popup?.destroy();
      root?.unmount();
      popup = null;
      root = null;
      container = null;
    },
  };
}
