import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useMessageSelection } from '../useMessageSelection';
import type { ChatMessage } from '../../types';

const makeMessages = (count: number): ChatMessage[] =>
  Array.from({ length: count }, (_, i) => ({
    id: i + 1,
    roomId: 1,
    senderId: i % 2 === 0 ? 1 : 2,
    senderName: `User${i % 2}`,
    content: `msg${i + 1}`,
    createdAt: '2025-01-01T00:00:00',
    isSender: i % 2 === 0,
  }));

describe('useMessageSelection', () => {
  it('初期状態はselectionMode=false', () => {
    const { result } = renderHook(() => useMessageSelection(makeMessages(5)));
    expect(result.current.selectionMode).toBe(false);
    expect(result.current.selectedMessages.size).toBe(0);
  });

  it('enterSelectionModeで選択モードに入る', () => {
    const { result } = renderHook(() => useMessageSelection(makeMessages(5)));
    act(() => result.current.enterSelectionMode());
    expect(result.current.selectionMode).toBe(true);
  });

  it('cancelSelectionで選択モードを終了する', () => {
    const { result } = renderHook(() => useMessageSelection(makeMessages(5)));
    act(() => result.current.enterSelectionMode());
    act(() => result.current.cancelSelection());
    expect(result.current.selectionMode).toBe(false);
    expect(result.current.selectedMessages.size).toBe(0);
  });

  it('handleRangeClickで開始・終了を指定して範囲選択する', () => {
    const messages = makeMessages(5);
    const { result } = renderHook(() => useMessageSelection(messages));
    act(() => result.current.enterSelectionMode());
    // 1番目をクリック（開始）
    act(() => result.current.handleRangeClick(1));
    expect(result.current.selectedMessages.size).toBe(1);
    // 3番目をクリック（終了）
    act(() => result.current.handleRangeClick(3));
    expect(result.current.selectedMessages.size).toBe(3);
    expect(result.current.selectedMessages.has(1)).toBe(true);
    expect(result.current.selectedMessages.has(2)).toBe(true);
    expect(result.current.selectedMessages.has(3)).toBe(true);
  });

  it('handleQuickSelectで直近N件を選択する', () => {
    const messages = makeMessages(10);
    const { result } = renderHook(() => useMessageSelection(messages));
    act(() => result.current.handleQuickSelect(3));
    expect(result.current.selectedMessages.size).toBe(3);
    expect(result.current.selectedMessages.has(8)).toBe(true);
    expect(result.current.selectedMessages.has(9)).toBe(true);
    expect(result.current.selectedMessages.has(10)).toBe(true);
  });

  it('handleSelectAllで全件選択する', () => {
    const messages = makeMessages(5);
    const { result } = renderHook(() => useMessageSelection(messages));
    act(() => result.current.handleSelectAll());
    expect(result.current.selectedMessages.size).toBe(5);
  });

  it('handleDeselectAllで全件選択解除する', () => {
    const messages = makeMessages(5);
    const { result } = renderHook(() => useMessageSelection(messages));
    act(() => result.current.handleSelectAll());
    act(() => result.current.handleDeselectAll());
    expect(result.current.selectedMessages.size).toBe(0);
  });

  it('isInRangeが範囲内のインデックスを判定する', () => {
    const messages = makeMessages(5);
    const { result } = renderHook(() => useMessageSelection(messages));
    act(() => result.current.handleRangeClick(1));
    act(() => result.current.handleRangeClick(3));
    expect(result.current.isInRange(0)).toBe(true);
    expect(result.current.isInRange(1)).toBe(true);
    expect(result.current.isInRange(2)).toBe(true);
    expect(result.current.isInRange(3)).toBe(false);
  });

  it('getRangeLabelが開始・終了ラベルを返す', () => {
    const messages = makeMessages(5);
    const { result } = renderHook(() => useMessageSelection(messages));
    act(() => result.current.handleRangeClick(1));
    act(() => result.current.handleRangeClick(3));
    expect(result.current.getRangeLabel(0)).toBe('開始');
    expect(result.current.getRangeLabel(2)).toBe('終了');
    expect(result.current.getRangeLabel(1)).toBeNull();
  });
});
