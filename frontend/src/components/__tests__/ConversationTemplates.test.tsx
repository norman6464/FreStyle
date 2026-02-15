import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ConversationTemplates from '../ConversationTemplates';

describe('ConversationTemplates', () => {
  const mockOnSelect = vi.fn();

  it('テンプレートカテゴリが全て表示される', () => {
    render(<ConversationTemplates onSelect={mockOnSelect} />);

    expect(screen.getByText('メール添削')).toBeInTheDocument();
    expect(screen.getByText('敬語・表現')).toBeInTheDocument();
    expect(screen.getByText('報連相')).toBeInTheDocument();
    expect(screen.getByText('会議・発表')).toBeInTheDocument();
  });

  it('テンプレートをクリックするとonSelectが呼ばれる', () => {
    render(<ConversationTemplates onSelect={mockOnSelect} />);

    const firstTemplate = screen.getAllByRole('button')[0];
    fireEvent.click(firstTemplate);

    expect(mockOnSelect).toHaveBeenCalledTimes(1);
    expect(mockOnSelect).toHaveBeenCalledWith(expect.any(String));
  });

  it('各テンプレートにタイトルと説明が表示される', () => {
    render(<ConversationTemplates onSelect={mockOnSelect} />);

    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThanOrEqual(8);

    buttons.forEach((button) => {
      expect(button.textContent).toBeTruthy();
    });
  });

  it('見出しテキストが表示される', () => {
    render(<ConversationTemplates onSelect={mockOnSelect} />);

    expect(screen.getByText('テンプレートから始める')).toBeInTheDocument();
  });
});
