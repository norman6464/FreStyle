import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SecondaryPanel from '../ui/SecondaryPanel';

describe('SecondaryPanel', () => {
  it('タイトルを表示する', () => {
    render(
      <SecondaryPanel title="チャット">
        <div>コンテンツ</div>
      </SecondaryPanel>
    );
    const titles = screen.getAllByText('チャット');
    expect(titles.length).toBeGreaterThanOrEqual(1);
  });

  it('子要素を表示する', () => {
    render(
      <SecondaryPanel title="テスト">
        <div>子要素コンテンツ</div>
      </SecondaryPanel>
    );
    const elements = screen.getAllByText('子要素コンテンツ');
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });

  it('ヘッダーコンテンツを表示する', () => {
    render(
      <SecondaryPanel title="テスト" headerContent={<input placeholder="検索" />}>
        <div>内容</div>
      </SecondaryPanel>
    );
    const inputs = screen.getAllByPlaceholderText('検索');
    expect(inputs.length).toBeGreaterThanOrEqual(1);
  });

  it('モバイル閉じるボタンを表示する', () => {
    render(
      <SecondaryPanel title="テスト" mobileOpen={true} onMobileClose={() => {}}>
        <div>内容</div>
      </SecondaryPanel>
    );
    expect(screen.getByLabelText('パネルを閉じる')).toBeDefined();
  });

  it('collapsible のとき折りたたみボタンを出し、クリックでトグルを呼ぶ', () => {
    const onToggle = vi.fn();
    render(
      <SecondaryPanel title="テスト" collapsible collapsed={false} onToggleCollapsed={onToggle}>
        <div>内容</div>
      </SecondaryPanel>
    );
    const btn = screen.getByLabelText('パネルを折りたたむ');
    fireEvent.click(btn);
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  it('折りたたみ中は「開く」ボタンを出し、折りたたみボタンは出さない', () => {
    const onToggle = vi.fn();
    render(
      <SecondaryPanel title="テスト" collapsible collapsed onToggleCollapsed={onToggle}>
        <div>章リスト内容</div>
      </SecondaryPanel>
    );
    // 折りたたみ中はデスクトップの全幅パネル（折りたたむボタン）を描画しない。
    expect(screen.queryByLabelText('パネルを折りたたむ')).not.toBeInTheDocument();
    fireEvent.click(screen.getByLabelText('パネルを開く'));
    expect(onToggle).toHaveBeenCalledTimes(1);
  });
});

describe('SecondaryPanel peekable（一時表示/固定表示）', () => {
  function createMockStorage(): Storage {
    let store: Record<string, string> = {};
    return {
      getItem: vi.fn((key: string) => store[key] ?? null),
      setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
      removeItem: vi.fn((key: string) => { delete store[key]; }),
      clear: vi.fn(() => { store = {}; }),
      get length() { return Object.keys(store).length; },
      key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
    };
  }

  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  const KEY = 'frestyle.panel.spec';

  function renderPeekable() {
    return render(
      <SecondaryPanel title="ノート" badge="3件" peekable storageKey={KEY}>
        <p>一覧の中身</p>
      </SecondaryPanel>,
    );
  }

  it('既定は固定表示で、«（サイドバーを閉じる）を出す', () => {
    renderPeekable();
    expect(screen.getAllByText('一覧の中身').length).toBeGreaterThan(0);
    expect(screen.getByRole('button', { name: 'サイドバーを閉じる' })).toBeInTheDocument();
  });

  it('« で一時表示モードになり、☰（固定表示する）が出て localStorage に保存される', () => {
    renderPeekable();
    fireEvent.click(screen.getByRole('button', { name: 'サイドバーを閉じる' }));
    expect(screen.getAllByRole('button', { name: 'サイドバーを固定表示する' }).length).toBeGreaterThan(0);
    expect(JSON.parse(localStorage.getItem(KEY)!)).toBe('collapsed');
  });

  it('☰ クリックで固定表示へ戻る', () => {
    localStorage.setItem(KEY, JSON.stringify('collapsed'));
    renderPeekable();
    fireEvent.click(screen.getAllByRole('button', { name: 'サイドバーを固定表示する' })[0]);
    expect(screen.getByRole('button', { name: 'サイドバーを閉じる' })).toBeInTheDocument();
    expect(JSON.parse(localStorage.getItem(KEY)!)).toBe('pinned');
  });

  it('一時表示モードでは ☰ ホバーでオーバーレイが浮く（中身が pointer-events を持つ）', () => {
    localStorage.setItem(KEY, JSON.stringify('collapsed'));
    const { container } = renderPeekable();
    const hamburger = screen.getAllByRole('button', { name: 'サイドバーを固定表示する' })[0];
    fireEvent.mouseEnter(hamburger);
    // オーバーレイ（translate-x-0）に切り替わる。
    const overlay = container.querySelector('.rounded-r-xl');
    expect(overlay?.className).toContain('translate-x-0');
    expect(overlay?.className).not.toContain('pointer-events-none');
  });

  it('保存済みモード（collapsed）で初期化される', () => {
    localStorage.setItem(KEY, JSON.stringify('collapsed'));
    renderPeekable();
    expect(screen.queryByRole('button', { name: 'サイドバーを閉じる' })).not.toBeInTheDocument();
  });
});
