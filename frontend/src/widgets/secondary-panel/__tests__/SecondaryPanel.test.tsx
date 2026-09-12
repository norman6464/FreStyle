import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import SecondaryPanel from '../ui/SecondaryPanel';

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

  it('side="right" のとき、モバイル固定パネルは右から出る（right-0・border-l を持ち、left-0 は持たない）', () => {
    render(
      <SecondaryPanel title="コメント" side="right" mobileOpen={true} onMobileClose={() => {}}>
        <div>内容</div>
      </SecondaryPanel>
    );
    // モバイル固定パネル自身を、閉じるボタン（アクセシブルな名前を持つ）から辿って探す
    // （querySelector だけだと配置クラスは検証できても、閉じる操作がアクセシブルな
    // 名前を持つことまでは検証できない — CodeRabbit 指摘）。
    const closeButton = screen.getByRole('button', { name: 'パネルを閉じる' });
    const mobilePanel = closeButton.closest('.inset-y-0');
    expect(mobilePanel).not.toBeNull();
    expect(mobilePanel?.className).toContain('right-0');
    expect(mobilePanel?.className).not.toContain('left-0');
    expect(mobilePanel?.className).toContain('border-l');
    expect(mobilePanel?.className).not.toContain('border-r');
  });

  it('side を省くと従来どおり左から出る（既定 "left"・後方互換）', () => {
    render(
      <SecondaryPanel title="ナレッジ" mobileOpen={true} onMobileClose={() => {}}>
        <div>内容</div>
      </SecondaryPanel>
    );
    const closeButton = screen.getByRole('button', { name: 'パネルを閉じる' });
    const mobilePanel = closeButton.closest('.inset-y-0');
    expect(mobilePanel?.className).toContain('left-0');
    expect(mobilePanel?.className).not.toContain('right-0');
    expect(mobilePanel?.className).toContain('border-r');
  });

  it.each([
    ['left（既定）', undefined],
    ['right', 'right' as const],
  ])('デスクトップの通常表示は本文との境目が見えるよう左右とも線を持つ（side=%s）', (_label, side) => {
    const { container } = render(
      <SecondaryPanel title="ナレッジ" side={side}>
        <div>内容</div>
      </SecondaryPanel>
    );
    // 呼び出し側の DOM 順（本文の左右どちらに置くか）だけで境目側が決まるよう、
    // side による分岐を持たず常に両側へ線を引く（左右どちらに置いても境目が見える）。
    const panel = container.querySelector('div.w-72.hidden.md\\:flex');
    expect(panel?.className).toContain('border-x');
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
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  const KEY = 'frestyle.panel.spec';

  function renderPeekable() {
    return render(
      <SecondaryPanel title="ナレッジ" badge="3件" peekable storageKey={KEY}>
        <p>一覧の中身</p>
      </SecondaryPanel>,
    );
  }

  it('既定は固定表示で、«（サイドバーを閉じる）を出す', () => {
    renderPeekable();
    expect(screen.getAllByText('一覧の中身').length).toBeGreaterThan(0);
    expect(screen.getByRole('button', { name: 'サイドバーを閉じる' })).toBeInTheDocument();
  });

  it('最初から固定表示のとき（再訪・再読み込み）は入場アニメーションを付けない', () => {
    const { container } = renderPeekable();
    const panel = screen.getByRole('button', { name: 'サイドバーを閉じる' }).closest('div.w-72');
    expect(panel?.className).not.toContain('animate-panel-pin');
    expect(container.querySelector('.animate-panel-pin')).toBeNull();
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

  it('一時表示から固定表示へ切り替えた瞬間だけ、入場アニメーションを付ける', () => {
    localStorage.setItem(KEY, JSON.stringify('collapsed'));
    renderPeekable();
    fireEvent.click(screen.getAllByRole('button', { name: 'サイドバーを固定表示する' })[0]);
    const panel = screen.getByRole('button', { name: 'サイドバーを閉じる' }).closest('div.w-72');
    expect(panel?.className).toContain('animate-panel-pin');
  });

  it('一時表示モードでは左端ホバーでオーバーレイが浮く（中身が pointer-events を持つ）', () => {
    localStorage.setItem(KEY, JSON.stringify('collapsed'));
    const { container } = renderPeekable();
    // 固定表示に戻すボタン（旧☰）はヘッダー側（widgets/app-shell/ui/Header.tsx）に
    // 移設済みで、ここ（本文側）に残るのは左端の透明なホバー検知ゾーンだけ。
    const edgeZone = container.querySelector('[aria-hidden="true"].w-2');
    expect(edgeZone).not.toBeNull();
    fireEvent.mouseEnter(edgeZone!);
    // オーバーレイ（translate-x-0）に切り替わる。
    const overlay = container.querySelector('.rounded-r-xl');
    expect(overlay?.className).toContain('translate-x-0');
    expect(overlay?.className).not.toContain('pointer-events-none');
  });

  it('固定表示に戻すハンバーガーは本文側に常時表示しない（ヘッダーへ移設済み）', () => {
    localStorage.setItem(KEY, JSON.stringify('collapsed'));
    renderPeekable();
    // 残るのはオーバーレイ内部の ☰（一時表示中のみ意味を持つ）だけ。常時見える入口は無い。
    expect(screen.getAllByRole('button', { name: 'サイドバーを固定表示する' })).toHaveLength(1);
  });

  it('保存済みモード（collapsed）で初期化される', () => {
    localStorage.setItem(KEY, JSON.stringify('collapsed'));
    renderPeekable();
    expect(screen.queryByRole('button', { name: 'サイドバーを閉じる' })).not.toBeInTheDocument();
  });
});

describe('SecondaryPanel resizable（縦線をドラッグして横幅を変える）', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
    vi.stubGlobal('innerWidth', 1200);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  function getHandleAndPanel() {
    const handle = screen.getByRole('separator', { name: 'サイドバーの幅を変更する' });
    return { handle, panel: handle.parentElement as HTMLElement };
  }

  it('resizable を指定しないとハンドルを出さない（既存画面への影響なし）', () => {
    render(
      <SecondaryPanel title="ナレッジ">
        <div>内容</div>
      </SecondaryPanel>
    );
    expect(screen.queryByRole('separator')).not.toBeInTheDocument();
  });

  it('resizable のとき、既定幅（288px）でハンドルを出す', () => {
    render(
      <SecondaryPanel title="ナレッジ" resizable>
        <div>内容</div>
      </SecondaryPanel>
    );
    const { panel } = getHandleAndPanel();
    expect(panel.style.width).toBe('288px');
  });

  it('defaultWidth を渡すとその幅で始まる', () => {
    render(
      <SecondaryPanel title="詳細" side="right" resizable defaultWidth={420}>
        <div>内容</div>
      </SecondaryPanel>
    );
    const { panel } = getHandleAndPanel();
    expect(panel.style.width).toBe('420px');
  });

  it('左側パネルは右へドラッグすると広がる', () => {
    render(
      <SecondaryPanel title="ナレッジ" resizable>
        <div>内容</div>
      </SecondaryPanel>
    );
    const { handle, panel } = getHandleAndPanel();
    fireEvent.mouseDown(handle, { clientX: 300 });
    fireEvent.mouseMove(document, { clientX: 350 });
    expect(panel.style.width).toBe('338px');
    fireEvent.mouseUp(document);
  });

  it('右側パネルは左へドラッグすると広がる（符号が逆）', () => {
    render(
      <SecondaryPanel title="詳細" side="right" resizable defaultWidth={420}>
        <div>内容</div>
      </SecondaryPanel>
    );
    const { handle, panel } = getHandleAndPanel();
    fireEvent.mouseDown(handle, { clientX: 500 });
    fireEvent.mouseMove(document, { clientX: 460 });
    expect(panel.style.width).toBe('460px');
    fireEvent.mouseUp(document);
  });

  it('画面幅の半分を超えては広がらない', () => {
    // innerWidth=1200 → 上限 600px
    render(
      <SecondaryPanel title="ナレッジ" resizable>
        <div>内容</div>
      </SecondaryPanel>
    );
    const { handle, panel } = getHandleAndPanel();
    fireEvent.mouseDown(handle, { clientX: 0 });
    fireEvent.mouseMove(document, { clientX: 900 });
    expect(panel.style.width).toBe('600px');
    fireEvent.mouseUp(document);
  });

  it('既定幅を下回っては狭まらない', () => {
    render(
      <SecondaryPanel title="ナレッジ" resizable>
        <div>内容</div>
      </SecondaryPanel>
    );
    const { handle, panel } = getHandleAndPanel();
    fireEvent.mouseDown(handle, { clientX: 300 });
    fireEvent.mouseMove(document, { clientX: -500 });
    expect(panel.style.width).toBe('288px');
    fireEvent.mouseUp(document);
  });

  it('矢印キーでも幅を変えられる（→ で広がる）', () => {
    render(
      <SecondaryPanel title="ナレッジ" resizable>
        <div>内容</div>
      </SecondaryPanel>
    );
    const { handle, panel } = getHandleAndPanel();
    fireEvent.keyDown(handle, { key: 'ArrowRight' });
    expect(panel.style.width).toBe('304px');
  });

  it('resizeStorageKey を渡すと、ドラッグ後の幅が localStorage に保存される', () => {
    render(
      <SecondaryPanel title="ナレッジ" resizable resizeStorageKey="test.panel.width">
        <div>内容</div>
      </SecondaryPanel>
    );
    const { handle } = getHandleAndPanel();
    fireEvent.mouseDown(handle, { clientX: 300 });
    fireEvent.mouseMove(document, { clientX: 350 });
    fireEvent.mouseUp(document);
    expect(JSON.parse(localStorage.getItem('test.panel.width')!)).toBe(338);
  });
});
