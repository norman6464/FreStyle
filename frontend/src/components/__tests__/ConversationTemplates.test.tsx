import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ConversationTemplates from '../ConversationTemplates';
import { CONVERSATION_TEMPLATES, TEMPLATE_CATEGORIES } from '../../constants/conversationTemplates';

describe('ConversationTemplates', () => {
  const mockOnSelect = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

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

  it('複数テンプレートをクリックしても各プロンプトが正しく渡される', () => {
    render(<ConversationTemplates onSelect={mockOnSelect} />);

    const buttons = screen.getAllByRole('button');
    fireEvent.click(buttons[0]);
    fireEvent.click(buttons[1]);

    expect(mockOnSelect).toHaveBeenCalledTimes(2);
    const firstPrompt = mockOnSelect.mock.calls[0][0];
    const secondPrompt = mockOnSelect.mock.calls[1][0];
    expect(firstPrompt).not.toBe(secondPrompt);
  });

  it('全カテゴリにアイコンが表示される', () => {
    const { container } = render(<ConversationTemplates onSelect={mockOnSelect} />);

    const categoryBadges = container.querySelectorAll('.rounded-full svg');
    expect(categoryBadges.length).toBe(TEMPLATE_CATEGORIES.length);
  });

  it('定数ファイルのテンプレート数と表示ボタン数が一致する', () => {
    render(<ConversationTemplates onSelect={mockOnSelect} />);

    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBe(CONVERSATION_TEMPLATES.length);
  });
});
