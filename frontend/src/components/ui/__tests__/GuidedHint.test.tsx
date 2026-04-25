import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import GuidedHint from '../GuidedHint';
import { createMockStorage } from '../../../test/mockStorage';

describe('GuidedHint', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('タイトルと本文を表示する', () => {
    render(
      <GuidedHint title="はじめての方へ">
        まずは練習モードから始めましょう。
      </GuidedHint>
    );
    expect(screen.getByText('はじめての方へ')).toBeInTheDocument();
    expect(screen.getByText(/まずは練習モード/)).toBeInTheDocument();
  });

  it('role="note" で描画される', () => {
    render(<GuidedHint title="t">body</GuidedHint>);
    expect(screen.getByRole('note')).toBeInTheDocument();
  });

  it('既定で閉じるボタンが表示される', () => {
    render(<GuidedHint title="t">body</GuidedHint>);
    expect(screen.getByRole('button', { name: 'ヒントを閉じる' })).toBeInTheDocument();
  });

  it('dismissible=false のとき閉じるボタンは表示されない', () => {
    render(
      <GuidedHint title="t" dismissible={false}>
        body
      </GuidedHint>
    );
    expect(screen.queryByRole('button', { name: 'ヒントを閉じる' })).not.toBeInTheDocument();
  });

  it('閉じるボタンを押すとヒントが消える', () => {
    render(<GuidedHint title="t">body</GuidedHint>);
    fireEvent.click(screen.getByRole('button', { name: 'ヒントを閉じる' }));
    expect(screen.queryByRole('note')).not.toBeInTheDocument();
  });

  it('storageKey 付きで閉じると localStorage に記録される', () => {
    render(
      <GuidedHint title="t" storageKey="hint:menu:first-visit">
        body
      </GuidedHint>
    );
    fireEvent.click(screen.getByRole('button', { name: 'ヒントを閉じる' }));
    expect(window.localStorage.getItem('hint:menu:first-visit')).toBe('dismissed');
  });

  it('既に dismissed が記録されていれば初回から非表示', () => {
    window.localStorage.setItem('hint:menu:first-visit', 'dismissed');
    render(
      <GuidedHint title="t" storageKey="hint:menu:first-visit">
        body
      </GuidedHint>
    );
    expect(screen.queryByRole('note')).not.toBeInTheDocument();
  });

  it('onDismiss が呼ばれる', () => {
    const onDismiss = vi.fn();
    render(
      <GuidedHint title="t" onDismiss={onDismiss}>
        body
      </GuidedHint>
    );
    fireEvent.click(screen.getByRole('button', { name: 'ヒントを閉じる' }));
    expect(onDismiss).toHaveBeenCalledOnce();
  });
});
