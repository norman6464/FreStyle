import { useState, useRef, useEffect } from 'react';
import { PlusIcon } from '@heroicons/react/24/outline';
import SlashCommandMenu from './SlashCommandMenu';
import { SLASH_COMMANDS, type SlashCommand } from '../constants/slashCommands';

interface BlockInserterButtonProps {
  visible: boolean;
  top: number;
  onCommand: (command: SlashCommand) => void;
}

export default function BlockInserterButton({ visible, top, onCommand }: BlockInserterButtonProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    setSelectedIndex(0);
  }, [menuOpen]);

  function handleSelect(index: number) {
    const cmd = SLASH_COMMANDS[index];
    if (cmd) {
      onCommand(cmd);
      setMenuOpen(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      setMenuOpen(false);
      return;
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex(prev => (prev + 1) % SLASH_COMMANDS.length);
      return;
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex(prev => (prev - 1 + SLASH_COMMANDS.length) % SLASH_COMMANDS.length);
      return;
    }
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSelect(selectedIndex);
    }
  }

  return (
    <div
      className="absolute left-0 z-10 transition-all duration-150"
      style={{ top: `${top}px` }}
    >
      <button
        type="button"
        aria-label="ブロックを追加"
        className={`w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] hover:text-[var(--color-text-secondary)] transition-opacity duration-150 ${
          visible || menuOpen ? 'opacity-100' : 'opacity-0'
        }`}
        onClick={() => setMenuOpen(!menuOpen)}
      >
        <PlusIcon className="w-4 h-4" />
      </button>
      {menuOpen && (
        <div
          ref={menuRef}
          className="absolute left-8 top-0 z-20"
          onKeyDown={handleKeyDown}
        >
          <SlashCommandMenu
            items={SLASH_COMMANDS}
            selectedIndex={selectedIndex}
            onSelect={handleSelect}
          />
        </div>
      )}
    </div>
  );
}
