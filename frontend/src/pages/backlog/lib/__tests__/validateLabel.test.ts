import { describe, it, expect } from 'vitest';
import { validateLabel, normalizeLabelColor, MAX_LABEL_NAME_LENGTH } from '../validateLabel';

describe('validateLabel', () => {
  it('妥当な名前と色は null（エラー無し）', () => {
    expect(validateLabel('不具合', '#1d4ed8')).toBeNull();
  });

  it('空文字・空白だけの名前は拒む', () => {
    expect(validateLabel('', '#1d4ed8')).toEqual({ name: 'ラベル名を入力してください。' });
    expect(validateLabel('   ', '#1d4ed8')).toEqual({ name: 'ラベル名を入力してください。' });
  });

  it(`名前が ${MAX_LABEL_NAME_LENGTH} 文字を超えたら拒む`, () => {
    const tooLong = 'あ'.repeat(MAX_LABEL_NAME_LENGTH + 1);
    expect(validateLabel(tooLong, '#1d4ed8')?.name).toBeDefined();
    expect(validateLabel('あ'.repeat(MAX_LABEL_NAME_LENGTH), '#1d4ed8')).toBeNull();
  });

  it('色の形が正しくなければ拒む', () => {
    for (const bad of ['', 'red', '#abc', '#gggggg', '1d4ed8']) {
      expect(validateLabel('不具合', bad)?.color).toBeDefined();
    }
  });

  it('大文字の16進も受け付ける（送信前に正規化する）', () => {
    expect(validateLabel('不具合', '#1D4ED8')).toBeNull();
  });

  it('名前と色の両方が不正なら両方のエラーを返す', () => {
    expect(validateLabel('', 'red')).toEqual({ name: 'ラベル名を入力してください。', color: '色を選んでください。' });
  });
});

describe('normalizeLabelColor', () => {
  it('小文字へ畳む', () => {
    expect(normalizeLabelColor('#1D4ED8')).toBe('#1d4ed8');
  });
});
