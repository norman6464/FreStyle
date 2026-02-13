import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ExportSessionButton from '../ExportSessionButton';

const mockMessages = [
  { id: 1, sessionId: 1, content: 'こんにちは', role: 'user' as const, isSender: true },
  { id: 2, sessionId: 1, content: 'こんにちは！何かお手伝いしましょうか？', role: 'assistant' as const, isSender: false },
  { id: 3, sessionId: 1, content: 'フィードバックをお願いします', role: 'user' as const, isSender: true },
];

describe('ExportSessionButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    });
  });

  it('ボタンが表示される', () => {
    render(<ExportSessionButton messages={mockMessages} />);

    expect(screen.getByTitle('会話をコピー')).toBeInTheDocument();
  });

  it('クリックでクリップボードにコピーされる', async () => {
    render(<ExportSessionButton messages={mockMessages} />);

    fireEvent.click(screen.getByTitle('会話をコピー'));

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        expect.stringContaining('こんにちは')
      );
    });
  });

  it('コピー後に完了メッセージが表示される', async () => {
    render(<ExportSessionButton messages={mockMessages} />);

    fireEvent.click(screen.getByTitle('会話をコピー'));

    await waitFor(() => {
      expect(screen.getByTitle('コピーしました')).toBeInTheDocument();
    });
  });

  it('メッセージが空の場合はボタンが無効になる', () => {
    render(<ExportSessionButton messages={[]} />);

    const button = screen.getByTitle('会話をコピー');
    expect(button).toBeDisabled();
  });
});
