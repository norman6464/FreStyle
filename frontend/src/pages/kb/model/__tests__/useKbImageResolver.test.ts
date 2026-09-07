import { act, renderHook } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useKbImageResolver } from '../useKbImageResolver';

const hoisted = vi.hoisted(() => ({
  issuePageImageDownloadURL: vi.fn(),
}));

vi.mock('@/entities/kb', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/kb')>();
  return {
    ...actual,
    KbRepository: {
      issuePageImageDownloadURL: hoisted.issuePageImageDownloadURL,
    },
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('useKbImageResolver', () => {
  it('"kb/" で始まらない src はそのまま返す（解決しに行かない）', async () => {
    const { result } = renderHook(() => useKbImageResolver('acme', 'p-1'));

    await expect(result.current.resolveImageSrc('https://cdn.example.com/a.png')).resolves.toBe(
      'https://cdn.example.com/a.png',
    );
    await expect(result.current.resolveImageSrc('data:image/png;base64,xx')).resolves.toBe(
      'data:image/png;base64,xx',
    );
    expect(hoisted.issuePageImageDownloadURL).not.toHaveBeenCalled();
  });

  it('"kb/" で始まる src は download-url を引いて解決後の URL を返す', async () => {
    hoisted.issuePageImageDownloadURL.mockResolvedValue({
      url: 'https://s3/signed?sig=1',
      expiresIn: 600,
    });
    const { result } = renderHook(() => useKbImageResolver('acme', 'p-1'));

    const got = await result.current.resolveImageSrc('kb/w-1/p-1/1.bin');

    expect(hoisted.issuePageImageDownloadURL).toHaveBeenCalledWith('acme', 'p-1', 'kb/w-1/p-1/1.bin');
    expect(got).toBe('https://s3/signed?sig=1');
  });

  it('同じ key を続けて解決しても、キャッシュが効いて 1 回しか引かない', async () => {
    hoisted.issuePageImageDownloadURL.mockResolvedValue({
      url: 'https://s3/signed?sig=1',
      expiresIn: 600,
    });
    const { result } = renderHook(() => useKbImageResolver('acme', 'p-1'));

    await result.current.resolveImageSrc('kb/w-1/p-1/1.bin');
    await result.current.resolveImageSrc('kb/w-1/p-1/1.bin');
    await result.current.resolveImageSrc('kb/w-1/p-1/1.bin');

    expect(hoisted.issuePageImageDownloadURL).toHaveBeenCalledTimes(1);
  });

  it('キャッシュの期限（9 分）を過ぎたら引き直す（10 分の期限より前倒しで切る）', async () => {
    vi.useFakeTimers();
    hoisted.issuePageImageDownloadURL
      .mockResolvedValueOnce({ url: 'https://s3/signed?sig=1', expiresIn: 600 })
      .mockResolvedValueOnce({ url: 'https://s3/signed?sig=2', expiresIn: 600 });
    const { result } = renderHook(() => useKbImageResolver('acme', 'p-1'));

    await act(async () => {
      await result.current.resolveImageSrc('kb/w-1/p-1/1.bin');
    });
    expect(hoisted.issuePageImageDownloadURL).toHaveBeenCalledTimes(1);

    // 9 分弱ならまだキャッシュのまま。
    await act(async () => {
      vi.advanceTimersByTime(8 * 60 * 1000 + 59_000);
    });
    await act(async () => {
      await result.current.resolveImageSrc('kb/w-1/p-1/1.bin');
    });
    expect(hoisted.issuePageImageDownloadURL).toHaveBeenCalledTimes(1);

    // 9 分を過ぎたら引き直す。
    await act(async () => {
      vi.advanceTimersByTime(2_000);
    });
    const got = await result.current.resolveImageSrc('kb/w-1/p-1/1.bin');
    expect(hoisted.issuePageImageDownloadURL).toHaveBeenCalledTimes(2);
    expect(got).toBe('https://s3/signed?sig=2');
  });

  it('解決に失敗したら例外を投げる（呼び出し側が失敗表示に倒す）', async () => {
    hoisted.issuePageImageDownloadURL.mockRejectedValue(new Error('404'));
    const { result } = renderHook(() => useKbImageResolver('acme', 'p-1'));

    await expect(result.current.resolveImageSrc('kb/w-1/p-1/missing.bin')).rejects.toThrow();
  });
});
