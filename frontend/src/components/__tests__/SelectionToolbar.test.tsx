import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import SelectionToolbar from '../SelectionToolbar';

const createMockChain = () => {
  const chain: Record<string, ReturnType<typeof vi.fn>> = {};
  const methods = ['focus', 'toggleBold', 'toggleItalic', 'toggleUnderline', 'toggleStrike', 'toggleCode', 'setParagraph', 'toggleHeading', 'setColor', 'unsetColor', 'setLink', 'unsetLink', 'run'];
  methods.forEach((m) => {
    chain[m] = vi.fn(() => chain);
  });
  return chain;
};

const createMockEditor = (overrides?: Partial<{ from: number; to: number; activeStates: Record<string, boolean> }>) => {
  const listeners: Record<string, ((...args: unknown[]) => void)[]> = {};
  const chain = createMockChain();
  const { from = 0, to = 5, activeStates = {} } = overrides || {};

  return {
    editor: {
      state: { selection: { from, to } },
      on: vi.fn((event: string, cb: (...args: unknown[]) => void) => {
        if (!listeners[event]) listeners[event] = [];
        listeners[event].push(cb);
      }),
      off: vi.fn(),
      isActive: vi.fn((type: string, attrs?: Record<string, unknown>) => {
        if (attrs && type === 'heading') return activeStates[`heading${attrs.level}`] || false;
        return activeStates[type] || false;
      }),
      chain: vi.fn(() => chain),
    },
    chain,
    listeners,
    emit: (event: string) => {
      listeners[event]?.forEach((cb) => cb());
    },
  };
};

describe('SelectionToolbar', () => {
  let originalGetSelection: typeof window.getSelection;
  let originalRAF: typeof window.requestAnimationFrame;
  let containerRef: { current: HTMLDivElement };

  beforeEach(() => {
    originalGetSelection = window.getSelection;
    originalRAF = window.requestAnimationFrame;

    window.requestAnimationFrame = (cb: FrameRequestCallback) => { cb(0); return 0; };

    const mockRange = {
      getBoundingClientRect: () => ({ top: 100, left: 200, width: 50, height: 20, bottom: 120, right: 250 }),
    };
    window.getSelection = vi.fn(() => ({
      rangeCount: 1,
      getRangeAt: () => mockRange,
    })) as unknown as typeof window.getSelection;

    const container = document.createElement('div');
    container.getBoundingClientRect = () => ({ top: 50, left: 50, width: 500, height: 400, bottom: 450, right: 550 } as DOMRect);
    containerRef = { current: container };
  });

  afterEach(() => {
    window.getSelection = originalGetSelection;
    window.requestAnimationFrame = originalRAF;
  });

  const renderWithSelection = (editorOverrides?: Parameters<typeof createMockEditor>[0]) => {
    const mock = createMockEditor(editorOverrides);
    render(<SelectionToolbar editor={mock.editor as never} containerRef={containerRef} />);

    act(() => {
      mock.emit('selectionUpdate');
    });

    return mock;
  };

  it('editor=nullの時は何も表示しない', () => {
    const { container } = render(<SelectionToolbar editor={null} containerRef={containerRef} />);
    expect(container.innerHTML).toBe('');
  });

  it('選択がない場合は表示されない', () => {
    const mock = createMockEditor({ from: 0, to: 0 });
    const { container } = render(<SelectionToolbar editor={mock.editor as never} containerRef={containerRef} />);

    act(() => { mock.emit('selectionUpdate'); });

    expect(container.innerHTML).toBe('');
  });

  it('選択がある場合にツールバーが表示される', () => {
    renderWithSelection();

    expect(screen.getByLabelText('太字')).toBeInTheDocument();
    expect(screen.getByLabelText('斜体')).toBeInTheDocument();
    expect(screen.getByLabelText('下線')).toBeInTheDocument();
    expect(screen.getByLabelText('取り消し線')).toBeInTheDocument();
    expect(screen.getByLabelText('インラインコード')).toBeInTheDocument();
    expect(screen.getByLabelText('リンク')).toBeInTheDocument();
    expect(screen.getByLabelText('文字色')).toBeInTheDocument();
  });

  it('太字ボタンでtoggleBoldが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByLabelText('太字'));

    expect(mock.chain.focus).toHaveBeenCalled();
    expect(mock.chain.toggleBold).toHaveBeenCalled();
    expect(mock.chain.run).toHaveBeenCalled();
  });

  it('斜体ボタンでtoggleItalicが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByLabelText('斜体'));

    expect(mock.chain.toggleItalic).toHaveBeenCalled();
  });

  it('下線ボタンでtoggleUnderlineが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByLabelText('下線'));

    expect(mock.chain.toggleUnderline).toHaveBeenCalled();
  });

  it('取り消し線ボタンでtoggleStrikeが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByLabelText('取り消し線'));

    expect(mock.chain.toggleStrike).toHaveBeenCalled();
  });

  it('コードボタンでtoggleCodeが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByLabelText('インラインコード'));

    expect(mock.chain.toggleCode).toHaveBeenCalled();
  });

  it('見出しドロップダウンが開閉できる', () => {
    renderWithSelection();

    expect(screen.queryByText('見出し1')).not.toBeInTheDocument();

    fireEvent.mouseDown(screen.getByText('テキスト'));

    expect(screen.getByText('見出し1')).toBeInTheDocument();
    expect(screen.getByText('見出し2')).toBeInTheDocument();
    expect(screen.getByText('見出し3')).toBeInTheDocument();
  });

  it('見出し選択でtoggleHeadingが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByText('テキスト'));
    fireEvent.mouseDown(screen.getByText('見出し2'));

    expect(mock.chain.toggleHeading).toHaveBeenCalledWith({ level: 2 });
    expect(mock.chain.run).toHaveBeenCalled();
  });

  it('テキスト選択でsetParagraphが実行される', () => {
    const mock = renderWithSelection({ activeStates: { heading2: true } });

    // 見出し2が現在のレベルとして表示される → クリックでドロップダウン開く
    fireEvent.mouseDown(screen.getByText('見出し2'));
    // ドロップダウン内の「テキスト」オプションをクリック
    fireEvent.mouseDown(screen.getByText('テキスト'));

    expect(mock.chain.setParagraph).toHaveBeenCalled();
  });

  it('色ドロップダウンが開閉できる', () => {
    renderWithSelection();

    expect(screen.queryByLabelText('赤')).not.toBeInTheDocument();

    fireEvent.mouseDown(screen.getByLabelText('文字色'));

    expect(screen.getByLabelText('赤')).toBeInTheDocument();
    expect(screen.getByLabelText('青')).toBeInTheDocument();
    expect(screen.getByLabelText('色をリセット')).toBeInTheDocument();
  });

  it('色選択でsetColorが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByLabelText('文字色'));
    fireEvent.mouseDown(screen.getByLabelText('赤'));

    expect(mock.chain.setColor).toHaveBeenCalledWith('#ef4444');
  });

  it('色リセットでunsetColorが実行される', () => {
    const mock = renderWithSelection();

    fireEvent.mouseDown(screen.getByLabelText('文字色'));
    fireEvent.mouseDown(screen.getByLabelText('色をリセット'));

    expect(mock.chain.unsetColor).toHaveBeenCalled();
  });

  it('見出しドロップダウンを開くと色ドロップダウンが閉じる', () => {
    renderWithSelection();

    // 色ドロップダウンを開く
    fireEvent.mouseDown(screen.getByLabelText('文字色'));
    expect(screen.getByLabelText('赤')).toBeInTheDocument();

    // 見出しドロップダウンを開く → 色が閉じる
    fireEvent.mouseDown(screen.getByText('テキスト'));

    expect(screen.queryByLabelText('赤')).not.toBeInTheDocument();
    expect(screen.getByText('見出し1')).toBeInTheDocument();
  });

  it('色ドロップダウンを開くと見出しドロップダウンが閉じる', () => {
    renderWithSelection();

    // 見出しドロップダウンを開く
    fireEvent.mouseDown(screen.getByText('テキスト'));
    expect(screen.getByText('見出し1')).toBeInTheDocument();

    // 色ドロップダウンを開く → 見出しが閉じる
    fireEvent.mouseDown(screen.getByLabelText('文字色'));

    expect(screen.queryByText('見出し1')).not.toBeInTheDocument();
    expect(screen.getByLabelText('赤')).toBeInTheDocument();
  });

  it('現在の見出しレベルが表示される', () => {
    renderWithSelection({ activeStates: { heading2: true } });

    expect(screen.getByText('見出し2')).toBeInTheDocument();
  });

  it('editorイベントリスナーが登録・解除される', () => {
    const mock = createMockEditor();
    const { unmount } = render(<SelectionToolbar editor={mock.editor as never} containerRef={containerRef} />);

    expect(mock.editor.on).toHaveBeenCalledWith('selectionUpdate', expect.any(Function));
    expect(mock.editor.on).toHaveBeenCalledWith('blur', expect.any(Function));

    unmount();

    expect(mock.editor.off).toHaveBeenCalledWith('selectionUpdate', expect.any(Function));
    expect(mock.editor.off).toHaveBeenCalledWith('blur', expect.any(Function));
  });
});
