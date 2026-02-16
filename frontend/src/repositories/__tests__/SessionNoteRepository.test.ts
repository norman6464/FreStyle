import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { SessionNoteRepository } from '../SessionNoteRepository';

const mockGet = vi.fn();
const mockPut = vi.fn();

vi.mock('../../lib/axios', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}));

function createMockStorage(): Storage {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { store = {}; }),
    get length() { return Object.keys(store).length; },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
}

describe('SessionNoteRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe('API正常系', () => {
    it('get: APIからメモを取得する', async () => {
      mockGet.mockResolvedValue({ data: { sessionId: 1, note: 'APIメモ', updatedAt: '2026-02-16T10:00:00' } });

      const result = await SessionNoteRepository.get(1);

      expect(result).not.toBeNull();
      expect(result!.note).toBe('APIメモ');
      expect(mockGet).toHaveBeenCalledWith('/api/session-notes/1');
    });

    it('save: APIでメモを保存する', async () => {
      mockPut.mockResolvedValue({});

      await SessionNoteRepository.save(1, '保存メモ');

      expect(mockPut).toHaveBeenCalledWith('/api/session-notes/1', { note: '保存メモ' });
    });
  });

  describe('APIエラー時のlocalStorageフォールバック', () => {
    it('get: APIエラー時はlocalStorageから取得する', async () => {
      mockGet.mockRejectedValue(new Error('Network Error'));
      localStorage.setItem('freestyle_session_notes', JSON.stringify({ '1': { sessionId: 1, note: 'ローカルメモ', updatedAt: '2026-02-16' } }));

      const result = await SessionNoteRepository.get(1);

      expect(result!.note).toBe('ローカルメモ');
    });

    it('get: APIエラー・localStorage空の場合はnullを返す', async () => {
      mockGet.mockRejectedValue(new Error('Network Error'));

      const result = await SessionNoteRepository.get(999);

      expect(result).toBeNull();
    });

    it('save: APIエラー時はlocalStorageに保存する', async () => {
      mockPut.mockRejectedValue(new Error('Network Error'));

      await SessionNoteRepository.save(1, 'フォールバックメモ');

      expect(localStorage.setItem).toHaveBeenCalled();
    });
  });

  describe('getAll', () => {
    it('localStorageから全メモを取得する', () => {
      localStorage.setItem('freestyle_session_notes', JSON.stringify({
        '1': { sessionId: 1, note: 'メモ1', updatedAt: '2026-02-16' },
        '2': { sessionId: 2, note: 'メモ2', updatedAt: '2026-02-16' },
      }));

      const result = SessionNoteRepository.getAll();

      expect(Object.keys(result)).toHaveLength(2);
      expect(result['1'].note).toBe('メモ1');
    });
  });
});
