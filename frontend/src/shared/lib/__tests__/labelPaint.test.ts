import { describe, it, expect } from 'vitest';
import { labelPaint } from '../labelPaint';

describe('labelPaint', () => {
  it('暗い色は地に敷いて白文字にする', () => {
    expect(labelPaint('#1d4ed8')).toEqual({ kind: 'solid', background: '#1d4ed8', color: '#ffffff' });
    expect(labelPaint('#5C5850')).toMatchObject({ kind: 'solid', color: '#ffffff' });
  });

  it('明るい色は地に敷いて濃い文字にする', () => {
    expect(labelPaint('#dbeafe')).toEqual({ kind: 'solid', background: '#dbeafe', color: '#191919' });
  });

  it('どちらの文字色でも基準に届かない中間の明るさは地に敷かない', () => {
    // 白との比も濃い文字との比も 4.5 に届かない色。地に敷くと必ず読みにくくなる。
    expect(labelPaint('#8b7355')).toEqual({ kind: 'outline', borderColor: '#8b7355' });
  });

  it('色の形が読めなければ枠線を既定の罫線に落とし、表示そのものは止めない', () => {
    for (const bad of ['', 'red', '#abc', '#12345', '#1d4ed8ff', 'rgb(0,0,0)']) {
      expect(labelPaint(bad)).toEqual({ kind: 'outline', borderColor: 'var(--color-surface-3)' });
    }
  });

  it('大文字の16進も読む', () => {
    expect(labelPaint('#1D4ED8')).toMatchObject({ kind: 'solid', color: '#ffffff' });
  });
});
