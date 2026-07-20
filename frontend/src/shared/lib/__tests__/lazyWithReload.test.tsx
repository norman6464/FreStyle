import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { Suspense } from 'react';
import { clearLazyReloadFlags, lazyWithReload } from '../lazyWithReload';

const reloadMock = vi.fn();
beforeEach(() => {
  vi.clearAllMocks();
  reloadMock.mockReset();
  // 各テストで sessionStorage を初期化
  window.sessionStorage.clear();
  // window.location.reload は jsdom で read-only なので Object.defineProperty で差替え
  Object.defineProperty(window, 'location', {
    value: { ...window.location, reload: reloadMock },
    writable: true,
  });
});

function FakeOk() {
  return <div>component-ok</div>;
}

describe('lazyWithReload', () => {
  it('正常時はラップしたコンポーネントを描画する', async () => {
    const Lazy = lazyWithReload(async () => ({ default: FakeOk }), 'ok-key');
    render(
      <Suspense fallback={<div>loading</div>}>
        <Lazy />
      </Suspense>
    );
    await waitFor(() => {
      expect(screen.getByText('component-ok')).toBeInTheDocument();
    });
    expect(reloadMock).not.toHaveBeenCalled();
  });

  it('chunk load 失敗で初回は window.location.reload を呼ぶ + sessionStorage flag をセット', async () => {
    const Lazy = lazyWithReload(async () => {
      throw new Error('Failed to fetch dynamically imported module: /assets/X-abc.js');
    }, 'chunk-fail');

    render(
      <Suspense fallback={<div>loading</div>}>
        <Lazy />
      </Suspense>
    );

    await waitFor(() => {
      expect(reloadMock).toHaveBeenCalledTimes(1);
    });
    expect(window.sessionStorage.getItem('lazy-reload:chunk-fail')).toBe('1');
  });

  it('reload flag が既にセットされている場合は再 reload せず元のエラーを投げる', async () => {
    const err = new Error('Failed to fetch dynamically imported module: /assets/X-abc.js');
    window.sessionStorage.setItem('lazy-reload:repeat', '1');

    const Lazy = lazyWithReload(async () => {
      throw err;
    }, 'repeat');

    // ErrorBoundary を簡易自作してキャッチする
    class Catcher extends (await import('react')).Component<
      { children: React.ReactNode },
      { caught: Error | null }
    > {
      state = { caught: null as Error | null };
      static getDerivedStateFromError(e: Error) {
        return { caught: e };
      }
      render() {
        if (this.state.caught) return <div>caught: {this.state.caught.message}</div>;
        return this.props.children;
      }
    }

    render(
      <Catcher>
        <Suspense fallback={<div>loading</div>}>
          <Lazy />
        </Suspense>
      </Catcher>
    );

    await waitFor(() => {
      expect(screen.getByText(/caught:/)).toBeInTheDocument();
    });
    expect(reloadMock).not.toHaveBeenCalled();
  });

  it('chunk load 以外のエラーは reload しない（普通のエラー扱い）', async () => {
    const Lazy = lazyWithReload(async () => {
      throw new Error('TypeError: x is undefined');
    }, 'other-err');

    class Catcher extends (await import('react')).Component<
      { children: React.ReactNode },
      { caught: Error | null }
    > {
      state = { caught: null as Error | null };
      static getDerivedStateFromError(e: Error) {
        return { caught: e };
      }
      render() {
        if (this.state.caught) return <div>err: {this.state.caught.message}</div>;
        return this.props.children;
      }
    }

    render(
      <Catcher>
        <Suspense fallback={<div>loading</div>}>
          <Lazy />
        </Suspense>
      </Catcher>
    );

    await waitFor(() => {
      expect(screen.getByText(/err:/)).toBeInTheDocument();
    });
    expect(reloadMock).not.toHaveBeenCalled();
  });

  it('clearLazyReloadFlags は lazy-reload: で始まる key を全部削除', () => {
    window.sessionStorage.setItem('lazy-reload:a', '1');
    window.sessionStorage.setItem('lazy-reload:b', '1');
    window.sessionStorage.setItem('other-key', 'keep');

    clearLazyReloadFlags();

    expect(window.sessionStorage.getItem('lazy-reload:a')).toBeNull();
    expect(window.sessionStorage.getItem('lazy-reload:b')).toBeNull();
    expect(window.sessionStorage.getItem('other-key')).toBe('keep');
  });
});
