import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { slashCommandRenderer } from '../slashCommandRenderer';
import type { SlashCommand } from '../../constants/slashCommands';

vi.mock('react-dom/client', () => ({
  createRoot: vi.fn(() => ({
    render: vi.fn(),
    unmount: vi.fn(),
  })),
}));

vi.mock('tippy.js', () => ({
  default: vi.fn(() => ({
    destroy: vi.fn(),
    hide: vi.fn(),
    setProps: vi.fn(),
  })),
}));

const mockItems: SlashCommand[] = [
  { label: 'テキスト', description: '通常の段落', icon: 'Bars3BottomLeftIcon', action: 'paragraph' },
  { label: '見出し1', description: '大見出し', icon: 'H1Icon', action: 'heading1' },
  { label: '見出し2', description: '中見出し', icon: 'H2Icon', action: 'heading2' },
];

function createMockProps(items = mockItems) {
  return {
    items,
    command: vi.fn(),
    clientRect: () => new DOMRect(0, 0, 100, 20),
    editor: {} as any,
    range: { from: 0, to: 0 },
    query: '',
    text: '',
    decorationNode: null,
  };
}

describe('slashCommandRenderer', () => {
  let renderer: ReturnType<typeof slashCommandRenderer>;

  beforeEach(() => {
    renderer = slashCommandRenderer();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('onStartでTippyポップアップを作成する', async () => {
    const tippy = (await import('tippy.js')).default;
    const props = createMockProps();
    renderer.onStart(props as any);

    expect(tippy).toHaveBeenCalledWith(
      document.body,
      expect.objectContaining({
        showOnCreate: true,
        interactive: true,
        trigger: 'manual',
        placement: 'bottom-start',
        theme: 'slash-command',
      })
    );
  });

  it('onStartでReactルートを作成してレンダリングする', async () => {
    const { createRoot } = await import('react-dom/client');
    const props = createMockProps();
    renderer.onStart(props as any);

    expect(createRoot).toHaveBeenCalled();
  });

  it('onUpdateでアイテムとポップアップ位置を更新する', async () => {
    const tippy = (await import('tippy.js')).default;
    const props = createMockProps();
    renderer.onStart(props as any);

    const mockPopup = vi.mocked(tippy).mock.results[0]!.value;
    const newProps = createMockProps([mockItems[0]!]);
    renderer.onUpdate(newProps as any);

    expect(mockPopup.setProps).toHaveBeenCalled();
  });

  it('ArrowDownキーでselectedIndexがインクリメントされる', () => {
    renderer.onStart(createMockProps() as any);

    const result = renderer.onKeyDown({
      event: new KeyboardEvent('keydown', { key: 'ArrowDown' }),
    } as any);

    expect(result).toBe(true);
  });

  it('ArrowUpキーでselectedIndexがデクリメントされる', () => {
    renderer.onStart(createMockProps() as any);

    const result = renderer.onKeyDown({
      event: new KeyboardEvent('keydown', { key: 'ArrowUp' }),
    } as any);

    expect(result).toBe(true);
  });

  it('Enterキーで選択中のコマンドが実行される', () => {
    const props = createMockProps();
    renderer.onStart(props as any);

    const result = renderer.onKeyDown({
      event: new KeyboardEvent('keydown', { key: 'Enter' }),
    } as any);

    expect(result).toBe(true);
    expect(props.command).toHaveBeenCalledWith(mockItems[0]);
  });

  it('Escapeキーでポップアップが非表示になる', async () => {
    const tippy = (await import('tippy.js')).default;
    renderer.onStart(createMockProps() as any);

    const mockPopup = vi.mocked(tippy).mock.results[0]!.value;
    const result = renderer.onKeyDown({
      event: new KeyboardEvent('keydown', { key: 'Escape' }),
    } as any);

    expect(result).toBe(true);
    expect(mockPopup.hide).toHaveBeenCalled();
  });

  it('未対応キーはfalseを返す', () => {
    renderer.onStart(createMockProps() as any);

    const result = renderer.onKeyDown({
      event: new KeyboardEvent('keydown', { key: 'Tab' }),
    } as any);

    expect(result).toBe(false);
  });

  it('onExitでポップアップとルートをクリーンアップする', async () => {
    const tippy = (await import('tippy.js')).default;
    const { createRoot } = await import('react-dom/client');
    renderer.onStart(createMockProps() as any);

    const mockPopup = vi.mocked(tippy).mock.results[0]!.value;
    const mockRoot = vi.mocked(createRoot).mock.results[0]!.value;

    renderer.onExit();

    expect(mockPopup.destroy).toHaveBeenCalled();
    expect(mockRoot.unmount).toHaveBeenCalled();
  });

  it('空アイテムでArrowDownを押してもfalseを返す', () => {
    renderer.onStart(createMockProps([]) as any);

    const result = renderer.onKeyDown({
      event: new KeyboardEvent('keydown', { key: 'ArrowDown' }),
    } as any);

    expect(result).toBe(false);
  });

  it('空アイテムでArrowUpを押してもfalseを返す', () => {
    renderer.onStart(createMockProps([]) as any);

    const result = renderer.onKeyDown({
      event: new KeyboardEvent('keydown', { key: 'ArrowUp' }),
    } as any);

    expect(result).toBe(false);
  });
});
