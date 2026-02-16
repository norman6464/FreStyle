import { useState, useMemo, useCallback } from 'react';
import { COMMAND_ITEMS, type CommandItem } from '../constants/commandPaletteItems';

function filterItems(items: CommandItem[], query: string): CommandItem[] {
  if (!query.trim()) return items;
  const lower = query.toLowerCase();
  return items.filter(item => {
    if (item.label.toLowerCase().includes(lower)) return true;
    if (item.description?.toLowerCase().includes(lower)) return true;
    if (item.keywords?.some(k => k.toLowerCase().includes(lower))) return true;
    return false;
  });
}

export function useCommandPalette() {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQueryState] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  const filteredItems = useMemo(() => filterItems(COMMAND_ITEMS, query), [query]);

  const setQuery = useCallback((q: string) => {
    setQueryState(q);
    setSelectedIndex(0);
  }, []);

  const open = useCallback(() => {
    setIsOpen(true);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
    setQueryState('');
    setSelectedIndex(0);
  }, []);

  const selectNext = useCallback(() => {
    setSelectedIndex(prev => {
      const items = filterItems(COMMAND_ITEMS, query);
      return (prev + 1) % items.length;
    });
  }, [query]);

  const selectPrev = useCallback(() => {
    setSelectedIndex(prev => {
      const items = filterItems(COMMAND_ITEMS, query);
      return (prev - 1 + items.length) % items.length;
    });
  }, [query]);

  return {
    isOpen,
    query,
    selectedIndex,
    filteredItems,
    open,
    close,
    setQuery,
    selectNext,
    selectPrev,
  };
}
