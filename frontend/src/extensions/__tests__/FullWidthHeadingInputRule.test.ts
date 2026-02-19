import { describe, it, expect } from 'vitest';

/**
 * 全角＃見出し変換の正規表現テスト
 *
 * TipTapのInputRuleで使用する正規表現が
 * 半角 # と全角 ＃ の両方にマッチすることを検証する。
 */

const HEADING_REGEX_1 = /^([#＃]{1})\s$/;
const HEADING_REGEX_2 = /^([#＃]{1,2})\s$/;
const HEADING_REGEX_3 = /^([#＃]{1,3})\s$/;

describe('全角＃見出し変換の正規表現', () => {
  it('半角 # + スペースでH1にマッチする', () => {
    expect(HEADING_REGEX_1.test('# ')).toBe(true);
  });

  it('全角 ＃ + スペースでH1にマッチする', () => {
    expect(HEADING_REGEX_1.test('＃ ')).toBe(true);
  });

  it('半角 ## + スペースでH2にマッチする', () => {
    expect(HEADING_REGEX_2.test('## ')).toBe(true);
  });

  it('全角 ＃＃ + スペースでH2にマッチする', () => {
    expect(HEADING_REGEX_2.test('＃＃ ')).toBe(true);
  });

  it('半角と全角混在 #＃ + スペースでH2にマッチする', () => {
    expect(HEADING_REGEX_2.test('#＃ ')).toBe(true);
  });

  it('半角 ### + スペースでH3にマッチする', () => {
    expect(HEADING_REGEX_3.test('### ')).toBe(true);
  });

  it('全角 ＃＃＃ + スペースでH3にマッチする', () => {
    expect(HEADING_REGEX_3.test('＃＃＃ ')).toBe(true);
  });

  it('4つ以上のハッシュにはH3がマッチしない', () => {
    expect(HEADING_REGEX_3.test('#### ')).toBe(false);
    expect(HEADING_REGEX_3.test('＃＃＃＃ ')).toBe(false);
  });
});

describe('全角＃ + Enter見出し変換', () => {
  // Enter押下時に使うパターン（スペース不要、ハッシュのみ）
  const HEADING_ENTER_PATTERN = /^[#＃]{1,3}$/;

  it('全角 ＃ のみでマッチする（H1）', () => {
    const match = '＃'.match(HEADING_ENTER_PATTERN);
    expect(match).not.toBeNull();
    expect(match![0].length).toBe(1);
  });

  it('全角 ＃＃ のみでマッチする（H2）', () => {
    const match = '＃＃'.match(HEADING_ENTER_PATTERN);
    expect(match).not.toBeNull();
    expect(match![0].length).toBe(2);
  });

  it('全角 ＃＃＃ のみでマッチする（H3）', () => {
    const match = '＃＃＃'.match(HEADING_ENTER_PATTERN);
    expect(match).not.toBeNull();
    expect(match![0].length).toBe(3);
  });

  it('半角 # でもマッチする', () => {
    expect(HEADING_ENTER_PATTERN.test('#')).toBe(true);
    expect(HEADING_ENTER_PATTERN.test('##')).toBe(true);
    expect(HEADING_ENTER_PATTERN.test('###')).toBe(true);
  });

  it('半角と全角混在でマッチする', () => {
    expect(HEADING_ENTER_PATTERN.test('#＃')).toBe(true);
    expect(HEADING_ENTER_PATTERN.test('＃#＃')).toBe(true);
  });

  it('4つ以上のハッシュにはマッチしない', () => {
    expect(HEADING_ENTER_PATTERN.test('####')).toBe(false);
    expect(HEADING_ENTER_PATTERN.test('＃＃＃＃')).toBe(false);
  });

  it('ハッシュ以外のテキストを含む場合はマッチしない', () => {
    expect(HEADING_ENTER_PATTERN.test('＃ テスト')).toBe(false);
    expect(HEADING_ENTER_PATTERN.test('テスト＃')).toBe(false);
    expect(HEADING_ENTER_PATTERN.test('＃ ')).toBe(false);
  });

  it('空文字にはマッチしない', () => {
    expect(HEADING_ENTER_PATTERN.test('')).toBe(false);
  });
});
