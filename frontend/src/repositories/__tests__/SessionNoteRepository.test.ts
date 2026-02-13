import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { SessionNoteRepository } from '../SessionNoteRepository';

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
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('初期状態ではnullを返す', () => {
    expect(SessionNoteRepository.get(1)).toBeNull();
  });

  it('メモを保存・取得できる', () => {
    SessionNoteRepository.save(1, '良い練習でした');
    const note = SessionNoteRepository.get(1);
    expect(note).not.toBeNull();
    expect(note!.note).toBe('良い練習でした');
    expect(note!.sessionId).toBe(1);
    expect(note!.updatedAt).toBeDefined();
  });

  it('メモを上書き保存できる', () => {
    SessionNoteRepository.save(1, '初回メモ');
    SessionNoteRepository.save(1, '更新メモ');
    const note = SessionNoteRepository.get(1);
    expect(note!.note).toBe('更新メモ');
  });

  it('異なるセッションのメモを独立して保存できる', () => {
    SessionNoteRepository.save(1, 'セッション1のメモ');
    SessionNoteRepository.save(2, 'セッション2のメモ');
    expect(SessionNoteRepository.get(1)!.note).toBe('セッション1のメモ');
    expect(SessionNoteRepository.get(2)!.note).toBe('セッション2のメモ');
  });
});
