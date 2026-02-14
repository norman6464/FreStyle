import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import RecentNotesCard from '../RecentNotesCard';

vi.mock('../../repositories/SessionNoteRepository', () => ({
  SessionNoteRepository: {
    getAll: vi.fn(),
  },
}));

import { SessionNoteRepository } from '../../repositories/SessionNoteRepository';

const mockGetAll = vi.mocked(SessionNoteRepository.getAll);

describe('RecentNotesCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('メモがない場合はnullを返す', () => {
    mockGetAll.mockReturnValue({});
    const { container } = render(<RecentNotesCard />);
    expect(container.firstChild).toBeNull();
  });

  it('タイトルが表示される', () => {
    mockGetAll.mockReturnValue({
      '1': { sessionId: 1, note: 'テストメモ', updatedAt: '2025-06-15T10:00:00Z' },
    });
    render(<RecentNotesCard />);
    expect(screen.getByText('最近のメモ')).toBeInTheDocument();
  });

  it('最新3件のメモが表示される', () => {
    mockGetAll.mockReturnValue({
      '1': { sessionId: 1, note: 'メモ1', updatedAt: '2025-06-13T10:00:00Z' },
      '2': { sessionId: 2, note: 'メモ2', updatedAt: '2025-06-14T10:00:00Z' },
      '3': { sessionId: 3, note: 'メモ3', updatedAt: '2025-06-15T10:00:00Z' },
      '4': { sessionId: 4, note: 'メモ4', updatedAt: '2025-06-12T10:00:00Z' },
    });
    render(<RecentNotesCard />);
    expect(screen.getByText('メモ3')).toBeInTheDocument();
    expect(screen.getByText('メモ2')).toBeInTheDocument();
    expect(screen.getByText('メモ1')).toBeInTheDocument();
    expect(screen.queryByText('メモ4')).not.toBeInTheDocument();
  });

  it('メモのプレビューが表示される', () => {
    mockGetAll.mockReturnValue({
      '1': { sessionId: 1, note: '振り返りの内容', updatedAt: '2025-06-15T10:00:00Z' },
    });
    render(<RecentNotesCard />);
    expect(screen.getByText('振り返りの内容')).toBeInTheDocument();
  });

  it('メモ件数が表示される', () => {
    mockGetAll.mockReturnValue({
      '1': { sessionId: 1, note: 'メモ1', updatedAt: '2025-06-15T10:00:00Z' },
      '2': { sessionId: 2, note: 'メモ2', updatedAt: '2025-06-14T10:00:00Z' },
    });
    render(<RecentNotesCard />);
    expect(screen.getByText('2件')).toBeInTheDocument();
  });
});
