import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ExportScoreHistoryButton from '../ExportScoreHistoryButton';
import type { ScoreHistoryItem } from '../../types';

// URL.createObjectURL / revokeObjectURL のモック
const mockCreateObjectURL = vi.fn(() => 'blob:mock-url');
const mockRevokeObjectURL = vi.fn();
global.URL.createObjectURL = mockCreateObjectURL;
global.URL.revokeObjectURL = mockRevokeObjectURL;

const history: ScoreHistoryItem[] = [
  {
    sessionId: 1,
    sessionTitle: '障害報告の練習',
    scenarioId: 3,
    overallScore: 7.5,
    scores: [
      { axis: '論理的構成力', score: 8, comment: '良い' },
      { axis: '配慮表現', score: 7, comment: 'まずまず' },
    ],
    createdAt: '2026-02-10T10:00:00',
  },
  {
    sessionId: 2,
    sessionTitle: '進捗報告',
    scenarioId: null,
    overallScore: 6.0,
    scores: [
      { axis: '論理的構成力', score: 6, comment: '' },
      { axis: '配慮表現', score: 6, comment: '' },
    ],
    createdAt: '2026-02-11T14:30:00',
  },
];

describe('ExportScoreHistoryButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.restoreAllMocks();
  });

  it('エクスポートボタンが表示される', () => {
    render(<ExportScoreHistoryButton history={history} />);

    expect(screen.getByRole('button', { name: /CSV/i })).toBeInTheDocument();
  });

  it('履歴が空の場合はボタンが無効になる', () => {
    render(<ExportScoreHistoryButton history={[]} />);

    expect(screen.getByRole('button', { name: /CSV/i })).toBeDisabled();
  });

  it('クリックでcreateObjectURLが呼ばれる', () => {
    render(<ExportScoreHistoryButton history={history} />);

    fireEvent.click(screen.getByRole('button', { name: /CSV/i }));

    expect(mockCreateObjectURL).toHaveBeenCalled();
    expect(mockRevokeObjectURL).toHaveBeenCalled();
  });

  it('クリックでBlobが作成される', () => {
    render(<ExportScoreHistoryButton history={history} />);

    fireEvent.click(screen.getByRole('button', { name: /CSV/i }));

    const blobArg = mockCreateObjectURL.mock.calls[0]?.[0];
    expect(blobArg).toBeInstanceOf(Blob);
  });
});
