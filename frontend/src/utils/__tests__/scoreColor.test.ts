import { describe, it, expect } from 'vitest';
import { getScoreTextColor, getScoreBarColor, getScoreLevel, getDeltaColor, formatDelta } from '../scoreColor';

describe('scoreColor', () => {
  describe('getScoreTextColor', () => {
    it('スコア8以上でemeraldを返す', () => {
      expect(getScoreTextColor(8)).toBe('text-emerald-400');
      expect(getScoreTextColor(10)).toBe('text-emerald-400');
    });

    it('スコア6以上8未満でamberを返す', () => {
      expect(getScoreTextColor(6)).toBe('text-amber-400');
      expect(getScoreTextColor(7.9)).toBe('text-amber-400');
    });

    it('スコア6未満でroseを返す', () => {
      expect(getScoreTextColor(5.9)).toBe('text-rose-400');
      expect(getScoreTextColor(0)).toBe('text-rose-400');
    });
  });

  describe('getScoreBarColor', () => {
    it('スコア8以上でemeraldを返す', () => {
      expect(getScoreBarColor(8)).toContain('emerald');
    });

    it('スコア6以上8未満でamberを返す', () => {
      expect(getScoreBarColor(6)).toContain('amber');
    });

    it('スコア6未満でroseを返す', () => {
      expect(getScoreBarColor(5)).toContain('rose');
    });
  });

  describe('getScoreLevel', () => {
    it('スコア8以上で優秀レベルを返す', () => {
      const result = getScoreLevel(8);
      expect(result.label).toBe('優秀レベル');
      expect(result.color).toContain('emerald');
    });

    it('スコア5以上8未満で実務レベルを返す', () => {
      const result = getScoreLevel(5);
      expect(result.label).toBe('実務レベル');
      expect(result.color).toContain('amber');
    });

    it('スコア5未満で基礎レベルを返す', () => {
      const result = getScoreLevel(4.9);
      expect(result.label).toBe('基礎レベル');
      expect(result.color).toContain('rose');
    });
  });

  describe('getDeltaColor', () => {
    it('正の値でemeraldを返す', () => {
      expect(getDeltaColor(0.5)).toBe('text-emerald-400');
    });

    it('負の値でroseを返す', () => {
      expect(getDeltaColor(-0.5)).toBe('text-rose-400');
    });

    it('0でfaintを返す', () => {
      expect(getDeltaColor(0)).toBe('text-[var(--color-text-faint)]');
    });
  });

  describe('formatDelta', () => {
    it('正の値で+を付ける', () => {
      expect(formatDelta(1.5)).toBe('+1.5');
    });

    it('負の値でそのまま表示する', () => {
      expect(formatDelta(-1.5)).toBe('-1.5');
    });

    it('0で±0を返す', () => {
      expect(formatDelta(0)).toBe('±0');
    });

    it('小数点以下1桁にフォーマットされる', () => {
      expect(formatDelta(1.23)).toBe('+1.2');
      expect(formatDelta(-0.78)).toBe('-0.8');
    });
  });

  describe('境界値テスト', () => {
    it('getScoreTextColor: 境界値7.999はamberを返す', () => {
      expect(getScoreTextColor(7.999)).toBe('text-amber-400');
    });

    it('getScoreTextColor: 境界値5.999はroseを返す', () => {
      expect(getScoreTextColor(5.999)).toBe('text-rose-400');
    });

    it('getScoreLevel: 境界値4.999は基礎レベルを返す', () => {
      expect(getScoreLevel(4.999).label).toBe('基礎レベル');
    });

    it('getDeltaColor: 非常に小さい正の値でemeraldを返す', () => {
      expect(getDeltaColor(0.001)).toBe('text-emerald-400');
    });
  });
});
