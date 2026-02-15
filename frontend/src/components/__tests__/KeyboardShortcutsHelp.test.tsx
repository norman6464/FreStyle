import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import KeyboardShortcutsHelp from '../KeyboardShortcutsHelp';

describe('KeyboardShortcutsHelp', () => {
  it('ボタンが表示される', () => {
    render(<KeyboardShortcutsHelp />);
    expect(screen.getByLabelText('ショートカット一覧')).toBeInTheDocument();
  });

  it('ボタンクリックでダイアログが表示される', () => {
    render(<KeyboardShortcutsHelp />);
    fireEvent.click(screen.getByLabelText('ショートカット一覧'));
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
  });

  it('ショートカット一覧に太字が含まれる', () => {
    render(<KeyboardShortcutsHelp />);
    fireEvent.click(screen.getByLabelText('ショートカット一覧'));
    expect(screen.getByText('太字')).toBeInTheDocument();
  });

  it('閉じるボタンでダイアログが閉じる', () => {
    render(<KeyboardShortcutsHelp />);
    fireEvent.click(screen.getByLabelText('ショートカット一覧'));
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();

    fireEvent.click(screen.getByLabelText('閉じる'));
    expect(screen.queryByText('キーボードショートカット')).not.toBeInTheDocument();
  });

  it('初期状態ではダイアログが非表示', () => {
    render(<KeyboardShortcutsHelp />);
    expect(screen.queryByText('キーボードショートカット')).not.toBeInTheDocument();
  });
});
