import { useState, useRef, useEffect } from 'react';
import { PlusIcon } from '@heroicons/react/24/outline';
import SlashCommandMenu from './SlashCommandMenu';
import { SLASH_COMMANDS, type SlashCommand } from '../constants/slashCommands';

interface BlockInserterButtonProps {
  visible: boolean;
  top: number;
  onCommand: (command: SlashCommand) => void;
  onMenuOpenChange?: (open: boolean) => void;
}

export default function BlockInserterButton({ visible, top, onCommand, onMenuOpenChange }: BlockInserterButtonProps) {
  const [menuOpen, setMenuOpenInternal] = useState(false);

  const setMenuOpen = (open: boolean) => {
    setMenuOpenInternal(open);
    onMenuOpenChange?.(open);
  };
  const [selectedIndex, setSelectedIndex] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    setSelectedIndex(0);
    menuRef.current?.focus();
  }, [menuOpen]);

  useEffect(() => {
    if (!menuOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [menuOpen]);

  useEffect(() => {
    if (!visible) {
      setMenuOpen(false);
    }
  }, [visible]);

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
      ref={containerRef}
      data-block-inserter
      className="absolute left-0 z-10 transition-all duration-150"
      style={{ top: `${top}px` }}
    >
      <div className="group relative">
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
        {!menuOpen && (
          <div className="absolute left-8 top-1/2 -translate-y-1/2 hidden group-hover:block z-30 pointer-events-none">
            <div className="bg-[var(--color-surface-1)] border border-surface-3 rounded-lg shadow-lg px-3 py-2 whitespace-nowrap">
              <p className="text-xs font-medium text-[var(--color-text-primary)]">クリックして下に追加</p>
              <p className="text-[11px] text-[var(--color-text-muted)]">Opt+クリック/Alt+クリックで上に追加</p>
            </div>
          </div>
        )}
      </div>
      {menuOpen && (
        <div
          ref={menuRef}
          tabIndex={0}
          className="absolute left-8 top-0 z-20 outline-none"
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
