import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import BlockInserterButton from '../BlockInserterButton';
import { SLASH_COMMANDS } from '../../constants/slashCommands';

describe('BlockInserterButton', () => {
  const mockOnCommand = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('「+」ボタンを表示する', () => {
    render(<BlockInserterButton visible={true} top={100} onCommand={mockOnCommand} />);
    expect(screen.getByLabelText('ブロックを追加')).toBeInTheDocument();
  });

  it('visible=falseの時は非表示', () => {
    render(<BlockInserterButton visible={false} top={100} onCommand={mockOnCommand} />);
    expect(screen.getByLabelText('ブロックを追加')).toHaveClass('opacity-0');
  });

  it('visible=trueの時は表示', () => {
    render(<BlockInserterButton visible={true} top={100} onCommand={mockOnCommand} />);
    expect(screen.getByLabelText('ブロックを追加')).toHaveClass('opacity-100');
  });

  it('クリックでメニューが表示される', () => {
    render(<BlockInserterButton visible={true} top={100} onCommand={mockOnCommand} />);
    fireEvent.click(screen.getByLabelText('ブロックを追加'));
    expect(screen.getByRole('menu')).toBeInTheDocument();
  });

  it('メニュー項目クリックでonCommandが呼ばれメニューが閉じる', () => {
    render(<BlockInserterButton visible={true} top={100} onCommand={mockOnCommand} />);
    fireEvent.click(screen.getByLabelText('ブロックを追加'));
    fireEvent.click(screen.getByText('見出し1'));
    expect(mockOnCommand).toHaveBeenCalledWith(SLASH_COMMANDS[1]);
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  it('topプロパティでボタン位置が設定される', () => {
    render(<BlockInserterButton visible={true} top={200} onCommand={mockOnCommand} />);
    const button = screen.getByLabelText('ブロックを追加');
    expect(button.parentElement).toHaveStyle({ top: '200px' });
  });

  it('メニュー表示中にEscで閉じる', () => {
    render(<BlockInserterButton visible={true} top={100} onCommand={mockOnCommand} />);
    fireEvent.click(screen.getByLabelText('ブロックを追加'));
    expect(screen.getByRole('menu')).toBeInTheDocument();
    fireEvent.keyDown(screen.getByRole('menu'), { key: 'Escape' });
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });
});
