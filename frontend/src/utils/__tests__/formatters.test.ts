import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { formatTime, truncateMessage } from '../formatters';

describe('formatTime', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2025-06-15T14:00:00'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('空文字の場合は空文字を返す', () => {
    expect(formatTime('')).toBe('');
  });

  it('今日の日付はHH:mm形式で返す', () => {
    const result = formatTime('2025-06-15T10:30:00');
    expect(result).toMatch(/10:30/);
  });

  it('昨日の場合は「昨日」を返す', () => {
    const result = formatTime('2025-06-14T10:30:00');
    expect(result).toBe('昨日');
  });

  it('1週間以内の場合は曜日を返す', () => {
    const result = formatTime('2025-06-12T10:30:00');
    expect(result).toMatch(/曜日/);
  });

  it('1週間以上前の場合は月日で返す', () => {
    const result = formatTime('2025-05-01T10:30:00');
    expect(result).toMatch(/5/);
  });
});

describe('truncateMessage', () => {
  it('undefinedの場合はデフォルトメッセージを返す', () => {
    expect(truncateMessage(undefined)).toBe('メッセージはありません');
  });

  it('短いメッセージはそのまま返す', () => {
    expect(truncateMessage('こんにちは')).toBe('こんにちは');
  });

  it('長いメッセージは省略される', () => {
    const long = 'あ'.repeat(50);
    const result = truncateMessage(long, 30);
    expect(result).toHaveLength(33); // 30 + '...'
    expect(result.endsWith('...')).toBe(true);
  });

  it('maxLengthを指定できる', () => {
    const result = truncateMessage('1234567890', 5);
    expect(result).toBe('12345...');
  });
});
