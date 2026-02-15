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

  it('選択中の項目にaria-current属性がある', () => {
    render(<SlashCommandMenu {...defaultProps} selectedIndex={2} />);
    const items = screen.getAllByRole('menuitem');
    expect(items[2]).toHaveAttribute('aria-current', 'true');
    expect(items[0]).toHaveAttribute('aria-current', 'false');
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
    expect(container.querySelector('[role="menu"]')?.children).toHaveLength(0);
  });

  it('menuロールを持つ', () => {
    render(<SlashCommandMenu {...defaultProps} />);
    expect(screen.getByRole('menu')).toBeInTheDocument();
  });

  it('aria-activedescendantが設定される', () => {
    render(<SlashCommandMenu {...defaultProps} selectedIndex={1} />);
    expect(screen.getByRole('menu')).toHaveAttribute('aria-activedescendant', 'slash-cmd-heading1');
  });

  it('aria-labelが設定される', () => {
    render(<SlashCommandMenu {...defaultProps} />);
    expect(screen.getByRole('menu')).toHaveAttribute('aria-label', 'スラッシュコマンド');
  });
});
