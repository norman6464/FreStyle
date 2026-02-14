import { describe, it, expect } from 'vitest';
import { getNoteStats } from '../noteStats';

describe('getNoteStats', () => {
  it('空文字で0文字・0分を返す', () => {
    const stats = getNoteStats('');
    expect(stats.charCount).toBe(0);
    expect(stats.readingTimeMin).toBe(0);
  });

  it('文字数を正しくカウントする', () => {
    const stats = getNoteStats('こんにちは世界');
    expect(stats.charCount).toBe(7);
  });

  it('改行を含む文字数を正しくカウントする', () => {
    const stats = getNoteStats('こんにちは\n世界');
    expect(stats.charCount).toBe(7);
  });

  it('読了時間を日本語400文字/分で計算する', () => {
    const text = 'あ'.repeat(400);
    const stats = getNoteStats(text);
    expect(stats.readingTimeMin).toBe(1);
  });

  it('800文字で2分の読了時間を返す', () => {
    const text = 'あ'.repeat(800);
    const stats = getNoteStats(text);
    expect(stats.readingTimeMin).toBe(2);
  });

  it('1文字でも最低1分の読了時間を返す', () => {
    const stats = getNoteStats('あ');
    expect(stats.readingTimeMin).toBe(1);
  });

  it('スペースを除外して文字数をカウントする', () => {
    const stats = getNoteStats('hello world');
    expect(stats.charCount).toBe(10);
  });

  it('全スペースの場合0文字・0分を返す', () => {
    const stats = getNoteStats('   \n\n  \t  ');
    expect(stats.charCount).toBe(0);
    expect(stats.readingTimeMin).toBe(0);
  });

  it('日本語と英語の混合文字をカウントする', () => {
    const stats = getNoteStats('Hello世界');
    expect(stats.charCount).toBe(7);
  });

  it('長文で正しい読了時間を計算する', () => {
    const text = 'あ'.repeat(1200);
    const stats = getNoteStats(text);
    expect(stats.readingTimeMin).toBe(3);
  });

  it('タブ文字を除外してカウントする', () => {
    const stats = getNoteStats('テスト\tデータ');
    expect(stats.charCount).toBe(6);
  });
});
