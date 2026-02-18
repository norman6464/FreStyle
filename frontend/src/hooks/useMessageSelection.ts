import { useState, useCallback } from 'react';
import type { ChatMessage } from '../types';

export function useMessageSelection(messages: ChatMessage[]) {
  const [selectionMode, setSelectionMode] = useState(false);
  const [selectedMessages, setSelectedMessages] = useState<Set<string>>(new Set());
  const [rangeStart, setRangeStart] = useState<number | null>(null);
  const [rangeEnd, setRangeEnd] = useState<number | null>(null);

  const enterSelectionMode = useCallback(() => {
    setSelectionMode(true);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  }, []);

  const cancelSelection = useCallback(() => {
    setSelectionMode(false);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  }, []);

  const handleRangeClick = useCallback((messageId: string) => {
    const messageIndex = messages.findIndex((msg) => msg.id === messageId);
    if (rangeStart === null) {
      setRangeStart(messageIndex);
      setRangeEnd(null);
      setSelectedMessages(new Set([messageId]));
    } else if (rangeEnd === null) {
      setRangeEnd(messageIndex);
      const start = Math.min(rangeStart, messageIndex);
      const end = Math.max(rangeStart, messageIndex);
      const rangeIds = new Set(messages.slice(start, end + 1).map((msg) => msg.id));
      setSelectedMessages(rangeIds);
    } else {
      setRangeStart(messageIndex);
      setRangeEnd(null);
      setSelectedMessages(new Set([messageId]));
    }
  }, [messages, rangeStart, rangeEnd]);

  const handleQuickSelect = useCallback((count: number) => {
    const recentMessages = messages.slice(-count);
    const recentIds = new Set(recentMessages.map((msg) => msg.id));
    setSelectedMessages(recentIds);
    if (recentMessages.length > 0) {
      setRangeStart(messages.length - count);
      setRangeEnd(messages.length - 1);
    }
  }, [messages]);

  const handleSelectAll = useCallback(() => {
    const allIds = new Set(messages.map((msg) => msg.id));
    setSelectedMessages(allIds);
    setRangeStart(0);
    setRangeEnd(messages.length - 1);
  }, [messages]);

  const handleDeselectAll = useCallback(() => {
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  }, []);

  const isInRange = useCallback((index: number): boolean => {
    if (rangeStart === null) return false;
    if (rangeEnd === null) return index === rangeStart;
    const start = Math.min(rangeStart, rangeEnd);
    const end = Math.max(rangeStart, rangeEnd);
    return index >= start && index <= end;
  }, [rangeStart, rangeEnd]);

  const getRangeLabel = useCallback((index: number): string | null => {
    if (rangeStart === index && rangeEnd === null) return '開始';
    if (rangeEnd === null) return null;
    const start = Math.min(rangeStart!, rangeEnd);
    const end = Math.max(rangeStart!, rangeEnd);
    if (index === start) return '開始';
    if (index === end) return '終了';
    return null;
  }, [rangeStart, rangeEnd]);

  return {
    selectionMode,
    selectedMessages,
    enterSelectionMode,
    cancelSelection,
    handleRangeClick,
    handleQuickSelect,
    handleSelectAll,
    handleDeselectAll,
    isInRange,
    getRangeLabel,
  };
}
