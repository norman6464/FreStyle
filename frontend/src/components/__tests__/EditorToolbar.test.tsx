import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import EditorToolbar from '../EditorToolbar';
import type { EditorFormatHandlers } from '../../hooks/useEditorFormat';

describe('EditorToolbar', () => {
  const createHandlers = (overrides?: Partial<EditorFormatHandlers>): EditorFormatHandlers => ({
    handleBold: vi.fn(),
    handleItalic: vi.fn(),
    handleUnderline: vi.fn(),
    handleStrike: vi.fn(),
    handleSuperscript: vi.fn(),
    handleSubscript: vi.fn(),
    handleSelectColor: vi.fn(),
    handleHighlight: vi.fn(),
    handleAlign: vi.fn(),
    handleUndo: vi.fn(),
    handleRedo: vi.fn(),
    handleClearFormatting: vi.fn(),
    handleIndent: vi.fn(),
    handleOutdent: vi.fn(),
    handleBlockquote: vi.fn(),
    handleHorizontalRule: vi.fn(),
    handleCodeBlock: vi.fn(),
    ...overrides,
  });

  it('書式ボタンが表示される', () => {
    render(<EditorToolbar handlers={createHandlers()} />);
    expect(screen.getByLabelText('太字')).toBeInTheDocument();
    expect(screen.getByLabelText('下線')).toBeInTheDocument();
  });

  it('カラーピッカーが表示される', () => {
    render(<EditorToolbar handlers={createHandlers()} />);
    expect(screen.getByLabelText('赤')).toBeInTheDocument();
  });

  it('配置ボタンが表示される', () => {
    render(<EditorToolbar handlers={createHandlers()} />);
    expect(screen.getByLabelText('左寄せ')).toBeInTheDocument();
    expect(screen.getByLabelText('中央寄せ')).toBeInTheDocument();
    expect(screen.getByLabelText('右寄せ')).toBeInTheDocument();
  });

  it('色選択でhandleSelectColorが呼ばれる', () => {
    const handleSelectColor = vi.fn();
    render(<EditorToolbar handlers={createHandlers({ handleSelectColor })} />);
    fireEvent.click(screen.getByLabelText('赤'));
    expect(handleSelectColor).toHaveBeenCalled();
  });

  it('配置ボタンクリックでhandleAlignが呼ばれる', () => {
    const handleAlign = vi.fn();
    render(<EditorToolbar handlers={createHandlers({ handleAlign })} />);
    fireEvent.click(screen.getByLabelText('中央寄せ'));
    expect(handleAlign).toHaveBeenCalledWith('center');
  });
});
