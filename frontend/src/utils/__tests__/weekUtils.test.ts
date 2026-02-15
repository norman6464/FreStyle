import { describe, it, expect } from 'vitest';
import { getMonday, getWeekRange } from '../weekUtils';

describe('weekUtils', () => {
  describe('getMonday', () => {
    it('水曜日の月曜日を返す', () => {
      const wed = new Date(2025, 0, 8); // 2025-01-08 水曜
      const monday = getMonday(wed);
      expect(monday.getDay()).toBe(1);
      expect(monday.getDate()).toBe(6);
    });

    it('月曜日なら同日を返す', () => {
      const mon = new Date(2025, 0, 6); // 2025-01-06 月曜
      const monday = getMonday(mon);
      expect(monday.getDate()).toBe(6);
    });

    it('日曜日なら前週の月曜を返す', () => {
      const sun = new Date(2025, 0, 12); // 2025-01-12 日曜
      const monday = getMonday(sun);
      expect(monday.getDate()).toBe(6);
    });

    it('時刻が00:00:00にリセットされる', () => {
      const date = new Date(2025, 0, 8, 15, 30, 45);
      const monday = getMonday(date);
      expect(monday.getHours()).toBe(0);
      expect(monday.getMinutes()).toBe(0);
      expect(monday.getSeconds()).toBe(0);
    });
  });

  describe('getWeekRange', () => {
    it('今週の範囲を返す', () => {
      const { start, end } = getWeekRange(0);
      expect(start.getDay()).toBe(1);
      expect(end.getTime() - start.getTime()).toBe(7 * 24 * 60 * 60 * 1000);
    });

    it('先週の範囲を返す', () => {
      const thisWeek = getWeekRange(0);
      const lastWeek = getWeekRange(1);
      expect(lastWeek.end.getTime()).toBe(thisWeek.start.getTime());
    });
  });
});
