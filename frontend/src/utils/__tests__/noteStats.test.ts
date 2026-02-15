import { describe, it, expect } from 'vitest';
import { getNoteStats } from '../noteStats';

describe('getNoteStats', () => {
  it('空文字で0文字を返す', () => {
    const stats = getNoteStats('');
    expect(stats.charCount).toBe(0);
  });

  it('文字数を正しくカウントする', () => {
    const stats = getNoteStats('こんにちは世界');
    expect(stats.charCount).toBe(7);
  });

  it('改行を含む文字数を正しくカウントする', () => {
    const stats = getNoteStats('こんにちは\n世界');
    expect(stats.charCount).toBe(7);
  });

  it('スペースを除外して文字数をカウントする', () => {
    const stats = getNoteStats('hello world');
    expect(stats.charCount).toBe(10);
  });

  it('全スペースの場合0文字を返す', () => {
    const stats = getNoteStats('   \n\n  \t  ');
    expect(stats.charCount).toBe(0);
  });

  it('日本語と英語の混合文字をカウントする', () => {
    const stats = getNoteStats('Hello世界');
    expect(stats.charCount).toBe(7);
  });

  it('タブ文字を除外してカウントする', () => {
    const stats = getNoteStats('テスト\tデータ');
    expect(stats.charCount).toBe(6);
  });
});
