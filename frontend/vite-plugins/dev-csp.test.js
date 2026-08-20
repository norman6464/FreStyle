import { describe, expect, it } from 'vitest';
import { addConnectSrc, devCsp } from './dev-csp';

const HTML = [
  '<!--',
  "  - connect-src 'self' https://api.frestyle.jp : 解説コメント（書き換えない）",
  '-->',
  '<meta http-equiv="Content-Security-Policy" content="default-src \'self\';',
  " connect-src 'self' https://api.frestyle.jp; object-src 'none'\" />",
].join('\n');

describe('addConnectSrc', () => {
  it('meta タグの connect-src に開発用オリジンを足す', () => {
    const result = addConnectSrc(HTML, ['http://localhost:8080']);

    expect(result).toContain(
      "connect-src 'self' https://api.frestyle.jp http://localhost:8080;",
    );
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
    expect(addConnectSrc(HTML, [undefined, ''])).toBe(HTML);
  });
});

describe('devCsp', () => {
  it('開発サーバ限定のプラグインを返す', () => {
    const plugin = devCsp(['http://localhost:8080']);

    expect(plugin.apply).toBe('serve');
    expect(plugin.transformIndexHtml(HTML)).toContain('http://localhost:8080');
  });
});
