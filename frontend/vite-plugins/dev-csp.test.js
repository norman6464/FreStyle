import { describe, expect, it } from 'vitest';
import { addConnectSrc, devCsp, toCspOrigin } from './dev-csp';

const HTML = [
  '<!--',
  "  - connect-src 'self' https://api.frestyle.jp : 解説コメント（書き換えない）",
  '-->',
  '<meta http-equiv="Content-Security-Policy" content="default-src \'self\';',
  " connect-src 'self' https://api.frestyle.jp; object-src 'none'\" />",
].join('\n');

describe('toCspOrigin', () => {
  it('origin だけを取り出す（path は落とす）', () => {
    expect(toCspOrigin('http://localhost:8080')).toBe('http://localhost:8080');
    expect(toCspOrigin('http://localhost:8080/api/v2')).toBe('http://localhost:8080');
    expect(toCspOrigin('https://api.example.com/base/')).toBe('https://api.example.com');
  });

  it('未設定なら null（CSP を変えない）', () => {
    expect(toCspOrigin('')).toBeNull();
    expect(toCspOrigin(undefined)).toBeNull();
  });

  it('CSP のディレクティブを増やせる値は弾く', () => {
    expect(() => toCspOrigin('http://localhost:8080; script-src *')).toThrow(
      /URL として解釈できません/,
    );
    expect(() => toCspOrigin('http://localhost:8080/x ; script-src *')).not.toThrow();
    expect(toCspOrigin('http://localhost:8080/x ; script-src *')).toBe('http://localhost:8080');
  });

  it('http / https 以外は弾く', () => {
    expect(() => toCspOrigin('ftp://example.com')).toThrow(/http \/ https のみ/);
    expect(() => toCspOrigin('javascript:alert(1)')).toThrow(/http \/ https のみ/);
  });

  it('URL として解釈できない値は弾く', () => {
    expect(() => toCspOrigin('localhost:8080')).toThrow(/http \/ https のみ/);
    expect(() => toCspOrigin('not a url')).toThrow(/URL として解釈できません/);
  });
});

describe('addConnectSrc', () => {
  it('meta タグの connect-src に開発用オリジンを足す', () => {
    const result = addConnectSrc(HTML, ['http://localhost:8080']);

    expect(result).toContain("connect-src 'self' https://api.frestyle.jp http://localhost:8080;");
  });

  it('解説コメント側の connect-src は書き換えない', () => {
    const result = addConnectSrc(HTML, ['http://localhost:8080']);

    expect(result).toContain(
      "  - connect-src 'self' https://api.frestyle.jp : 解説コメント（書き換えない）",
    );
  });

  it('他のディレクティブを壊さない', () => {
    const result = addConnectSrc(HTML, ['http://localhost:8080']);

    expect(result).toContain("default-src 'self';");
    expect(result).toContain("object-src 'none'");
  });

  it('追加するオリジンが無ければ HTML を変えない', () => {
    expect(addConnectSrc(HTML, [])).toBe(HTML);
    expect(addConnectSrc(HTML, [null, ''])).toBe(HTML);
  });
});

describe('devCsp', () => {
  it('開発サーバ限定のプラグインを返す', () => {
    const plugin = devCsp('http://localhost:8080/api');

    expect(plugin.apply).toBe('serve');
    expect(plugin.transformIndexHtml(HTML)).toContain('http://localhost:8080;');
  });

  it('未設定なら HTML を変えない', () => {
    expect(devCsp('').transformIndexHtml(HTML)).toBe(HTML);
  });

  it('不正な値は設定読み込み時に落とす', () => {
    expect(() => devCsp('ftp://example.com')).toThrow(/http \/ https のみ/);
  });
});
