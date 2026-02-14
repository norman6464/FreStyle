import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useCopyToClipboard } from '../useCopyToClipboard';

describe('useCopyToClipboard', () => {
  beforeEach(() => {
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    });
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('初期状態でcopiedIdはnull', () => {
    const { result } = renderHook(() => useCopyToClipboard());
    expect(result.current.copiedId).toBeNull();
  });

  it('copyToClipboardでクリップボードに書き込む', async () => {
    const { result } = renderHook(() => useCopyToClipboard());

    await act(async () => {
      await result.current.copyToClipboard(1, 'テストメッセージ');
    });

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('テストメッセージ');
  });

  it('コピー成功後にcopiedIdが設定される', async () => {
    const { result } = renderHook(() => useCopyToClipboard());

    await act(async () => {
      await result.current.copyToClipboard(42, 'テスト');
    });

    expect(result.current.copiedId).toBe(42);
  });

  it('コピー後2秒でcopiedIdがnullに戻る', async () => {
    const { result } = renderHook(() => useCopyToClipboard());

    await act(async () => {
      await result.current.copyToClipboard(1, 'テスト');
    });

    expect(result.current.copiedId).toBe(1);

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    expect(result.current.copiedId).toBeNull();
  });

  it('クリップボードAPI失敗時にcopiedIdが設定されない', async () => {
    vi.mocked(navigator.clipboard.writeText).mockRejectedValue(new Error('失敗'));
    const { result } = renderHook(() => useCopyToClipboard());

    await act(async () => {
      await result.current.copyToClipboard(1, 'テスト');
    });

    expect(result.current.copiedId).toBeNull();
  });

  it('連続コピーで前のタイマーがクリアされる', async () => {
    const { result } = renderHook(() => useCopyToClipboard());

    await act(async () => {
      await result.current.copyToClipboard(1, 'テスト1');
    });

    expect(result.current.copiedId).toBe(1);

    await act(async () => {
      await result.current.copyToClipboard(2, 'テスト2');
    });

    expect(result.current.copiedId).toBe(2);

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    expect(result.current.copiedId).toBeNull();
  });
});
