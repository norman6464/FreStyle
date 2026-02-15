import { describe, it, expect } from 'vitest';
import { toLocalDateKey, getCalendarDays, getIntensityClass } from '../calendarHelpers';

describe('calendarHelpers', () => {
  describe('toLocalDateKey', () => {
    it('日付をYYYY-MM-DD形式に変換する', () => {
      const date = new Date(2026, 0, 5); // 2026-01-05
      expect(toLocalDateKey(date)).toBe('2026-01-05');
    });

    it('月と日を0埋めする', () => {
      const date = new Date(2026, 2, 3); // 2026-03-03
      expect(toLocalDateKey(date)).toBe('2026-03-03');
    });
  });

  describe('getCalendarDays', () => {
    it('指定された週数分の日付を返す', () => {
      const days = getCalendarDays(1);
      expect(days).toHaveLength(7);
    });

    it('2週間分で14日を返す', () => {
      const days = getCalendarDays(2);
      expect(days).toHaveLength(14);
    });

    it('日付が昇順に並ぶ', () => {
      const days = getCalendarDays(2);
      for (let i = 1; i < days.length; i++) {
        expect(days[i].getTime()).toBeGreaterThan(days[i - 1].getTime());
      }
    });
  });

  describe('getIntensityClass', () => {
    it('0回でbg-surface-3を返す', () => {
      expect(getIntensityClass(0)).toBe('bg-surface-3');
    });

    it('1回でbg-emerald-200を返す', () => {
      expect(getIntensityClass(1)).toBe('bg-emerald-200');
    });

    it('2回でbg-emerald-400を返す', () => {
      expect(getIntensityClass(2)).toBe('bg-emerald-400');
    });

    it('3回以上でbg-emerald-600を返す', () => {
      expect(getIntensityClass(3)).toBe('bg-emerald-600');
      expect(getIntensityClass(5)).toBe('bg-emerald-600');
    });
  });
});
