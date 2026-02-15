import { describe, it, expect } from 'vitest';
import {
  polarToCartesian,
  getPolygonPoints,
  getGridPoints,
  RADAR_SIZE,
  RADAR_CENTER,
  RADAR_RADIUS,
  RADAR_GRID_LEVELS,
} from '../skillRadarHelpers';

describe('skillRadarHelpers', () => {
  describe('定数', () => {
    it('RADAR_SIZEは200', () => {
      expect(RADAR_SIZE).toBe(200);
    });

    it('RADAR_CENTERはSIZE/2', () => {
      expect(RADAR_CENTER).toBe(RADAR_SIZE / 2);
    });

    it('RADAR_RADIUSは70', () => {
      expect(RADAR_RADIUS).toBe(70);
    });

    it('RADAR_GRID_LEVELSは[2,4,6,8,10]', () => {
      expect(RADAR_GRID_LEVELS).toEqual([2, 4, 6, 8, 10]);
    });
  });

  describe('polarToCartesian', () => {
    it('角度0で上方向（y=CENTER-radius）を返す', () => {
      const point = polarToCartesian(0, RADAR_RADIUS);
      expect(point.x).toBeCloseTo(RADAR_CENTER, 5);
      expect(point.y).toBeCloseTo(RADAR_CENTER - RADAR_RADIUS, 5);
    });

    it('角度90で右方向（x=CENTER+radius）を返す', () => {
      const point = polarToCartesian(90, RADAR_RADIUS);
      expect(point.x).toBeCloseTo(RADAR_CENTER + RADAR_RADIUS, 5);
      expect(point.y).toBeCloseTo(RADAR_CENTER, 5);
    });

    it('半径0で中心を返す', () => {
      const point = polarToCartesian(45, 0);
      expect(point.x).toBeCloseTo(RADAR_CENTER, 5);
      expect(point.y).toBeCloseTo(RADAR_CENTER, 5);
    });
  });

  describe('getPolygonPoints', () => {
    it('空配列で空文字を返す', () => {
      expect(getPolygonPoints([], 10)).toBe('');
    });

    it('全て最大値で外周のポイントを返す', () => {
      const result = getPolygonPoints([10, 10, 10, 10, 10], 10);
      expect(result).toBeTruthy();
      const points = result.split(' ');
      expect(points).toHaveLength(5);
    });

    it('値0で全て中心を返す', () => {
      const result = getPolygonPoints([0, 0, 0], 10);
      const points = result.split(' ');
      points.forEach((point) => {
        const [x, y] = point.split(',').map(Number);
        expect(x).toBeCloseTo(RADAR_CENTER, 5);
        expect(y).toBeCloseTo(RADAR_CENTER, 5);
      });
    });
  });

  describe('getGridPoints', () => {
    it('count=0で空文字を返す', () => {
      expect(getGridPoints(5, 0)).toBe('');
    });

    it('レベル10で外周のポイントを返す', () => {
      const result = getGridPoints(10, 5);
      expect(result).toBeTruthy();
      const points = result.split(' ');
      expect(points).toHaveLength(5);
    });

    it('レベル5で半径50%のポイントを返す', () => {
      const result = getGridPoints(5, 5);
      const points = result.split(' ');
      // 最初のポイント（角度0=上方向）
      const [x, y] = points[0].split(',').map(Number);
      expect(x).toBeCloseTo(RADAR_CENTER, 5);
      expect(y).toBeCloseTo(RADAR_CENTER - RADAR_RADIUS * 0.5, 5);
    });
  });
});
