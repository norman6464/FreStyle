import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SessionNoteEditor from '../SessionNoteEditor';

const mockSaveNote = vi.fn();

vi.mock('../../hooks/useSessionNote', () => ({
  useSessionNote: () => ({
    note: mockNote,
    saveNote: mockSaveNote,
  }),
}));

let mockNote = '';

describe('SessionNoteEditor', () => {
  it('メモ入力欄が表示される', () => {
    render(<SessionNoteEditor sessionId={1} />);

    expect(screen.getByPlaceholderText(/振り返りメモ/)).toBeInTheDocument();
  });

  it('保存ボタンでsaveNoteが呼ばれる', () => {
    render(<SessionNoteEditor sessionId={1} />);

    const textarea = screen.getByPlaceholderText(/振り返りメモ/);
    fireEvent.change(textarea, { target: { value: '学びメモ' } });
    fireEvent.click(screen.getByText('保存'));

    expect(mockSaveNote).toHaveBeenCalledWith('学びメモ');
  });

  it('既存メモが表示される', () => {
    mockNote = '既存のメモ';

    render(<SessionNoteEditor sessionId={1} />);

    expect(screen.getByDisplayValue('既存のメモ')).toBeInTheDocument();

    mockNote = '';
  });
});
