import { describe, it, expect } from 'vitest';
import { parseErrorLines } from '../executionErrors';

describe('parseErrorLines', () => {
  it('go のコンパイルエラーから行番号を抽出する', () => {
    const stderr = `# command-line-arguments
./main.go:7:9: undefined: undefinedFn
./main.go:10:2: missing return`;
    const markers = parseErrorLines(stderr, 'go');
    expect(markers.map((m) => m.line)).toEqual([7, 10]);
    expect(markers[0].message).toContain('undefined: undefinedFn');
  });

  it('go の同一行の重複は最初のメッセージにまとめる', () => {
    const stderr = `./main.go:5:1: first
./main.go:5:9: second`;
    const markers = parseErrorLines(stderr, 'go');
    expect(markers).toHaveLength(1);
    expect(markers[0].line).toBe(5);
    expect(markers[0].message).toContain('first');
  });

  it('javascript の構文エラー(スタックトレース形式)から行番号と例外を抽出する', () => {
    const stderr = `./main.js:3
console.log("unclosed
            ^^^^^^^^^

SyntaxError: Invalid or unexpected token
    at wrapSafe (node:internal/modules/cjs/loader:1385:18)`;
    const markers = parseErrorLines(stderr, 'javascript');
    expect(markers.map((m) => m.line)).toEqual([3]);
    expect(markers[0].message).toContain('SyntaxError: Invalid or unexpected token');
  });

  it('typescript の実行エラーから行番号を抽出する', () => {
    const stderr = `./main.ts:5
throw new Error("boom");
^

Error: boom
    at Object.<anonymous> (./main.ts:5:7)`;
    const markers = parseErrorLines(stderr, 'typescript');
    expect(markers.map((m) => m.line)).toEqual([5]);
    expect(markers[0].message).toContain('Error: boom');
  });

  it('php の parse error から行番号を抽出する', () => {
    const stderr =
      'PHP Parse error:  syntax error, unexpected end of file in /tmp/code_abc.php on line 4';
    const markers = parseErrorLines(stderr, 'php');
    expect(markers.map((m) => m.line)).toEqual([4]);
    expect(markers[0].message).toContain('syntax error');
  });

  it('bash のエラーから行番号を抽出する', () => {
    const stderr = '/tmp/bash-exec-1/script.sh: line 3: foo: command not found';
    const markers = parseErrorLines(stderr, 'bash');
    expect(markers.map((m) => m.line)).toEqual([3]);
    expect(markers[0].message).toContain('command not found');
  });

  it('行番号を含まない stderr や未対応言語は空配列', () => {
    expect(parseErrorLines('some generic failure', 'go')).toEqual([]);
    expect(parseErrorLines('./main.go:7:9: x', 'docker')).toEqual([]);
    expect(parseErrorLines('', 'go')).toEqual([]);
  });

  it('行番号は昇順に整列される', () => {
    const stderr = `./main.go:9:1: b
./main.go:2:1: a`;
    expect(parseErrorLines(stderr, 'go').map((m) => m.line)).toEqual([2, 9]);
  });
});
