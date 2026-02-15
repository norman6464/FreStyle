import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SlashCommandMenu from '../SlashCommandMenu';
import { SLASH_COMMANDS } from '../../constants/slashCommands';

describe('SlashCommandMenu', () => {
  const defaultProps = {
    items: SLASH_COMMANDS,
    selectedIndex: 0,
    onSelect: vi.fn(),
  };

  it('全コマンドを表示する', () => {
    render(<SlashCommandMenu {...defaultProps} />);
    for (const cmd of SLASH_COMMANDS) {
      expect(screen.getByText(cmd.label)).toBeInTheDocument();
      expect(screen.getByText(cmd.description)).toBeInTheDocument();
    }
  });

  it('選択中の項目にaria-selected属性がある', () => {
    render(<SlashCommandMenu {...defaultProps} selectedIndex={2} />);
    const items = screen.getAllByRole('option');
    expect(items[2]).toHaveAttribute('aria-selected', 'true');
    expect(items[0]).toHaveAttribute('aria-selected', 'false');
  });

  it('項目クリックでonSelectが呼ばれる', () => {
    const onSelect = vi.fn();
    render(<SlashCommandMenu {...defaultProps} onSelect={onSelect} />);
    fireEvent.click(screen.getByText('見出し1'));
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it('コマンドアイコンを表示する', () => {
    render(<SlashCommandMenu {...defaultProps} />);
    expect(screen.getByText('H1')).toBeInTheDocument();
    expect(screen.getByText('H2')).toBeInTheDocument();
  });

  it('空のitems配列で何も表示しない', () => {
    const { container } = render(
      <SlashCommandMenu items={[]} selectedIndex={0} onSelect={vi.fn()} />
    );
    expect(container.querySelector('[role="listbox"]')?.children).toHaveLength(0);
  });

  it('listboxロールを持つ', () => {
    render(<SlashCommandMenu {...defaultProps} />);
    expect(screen.getByRole('listbox')).toBeInTheDocument();
  });
});
