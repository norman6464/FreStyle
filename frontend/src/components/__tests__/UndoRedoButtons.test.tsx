import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import UndoRedoButtons from '../UndoRedoButtons';

describe('UndoRedoButtons', () => {
  it('2つのボタンが表示される', () => {
    render(<UndoRedoButtons onUndo={vi.fn()} onRedo={vi.fn()} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(2);
  });

  it('元に戻すボタンにaria-labelがある', () => {
    render(<UndoRedoButtons onUndo={vi.fn()} onRedo={vi.fn()} />);
    expect(screen.getByLabelText('元に戻す')).toBeInTheDocument();
  });

  it('やり直すボタンにaria-labelがある', () => {
    render(<UndoRedoButtons onUndo={vi.fn()} onRedo={vi.fn()} />);
    expect(screen.getByLabelText('やり直す')).toBeInTheDocument();
  });

  it('元に戻すクリックでonUndoが呼ばれる', () => {
    const onUndo = vi.fn();
    render(<UndoRedoButtons onUndo={onUndo} onRedo={vi.fn()} />);
    fireEvent.click(screen.getByLabelText('元に戻す'));
    expect(onUndo).toHaveBeenCalled();
  });

  it('やり直すクリックでonRedoが呼ばれる', () => {
    const onRedo = vi.fn();
    render(<UndoRedoButtons onUndo={vi.fn()} onRedo={onRedo} />);
    fireEvent.click(screen.getByLabelText('やり直す'));
    expect(onRedo).toHaveBeenCalled();
  });
});
