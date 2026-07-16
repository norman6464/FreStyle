import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ExecutionResultTable from '../ExecutionResultTable';
import { CodeExecutionResult } from '../../../types';

const expected = '{"id":1}';

function result(over: Partial<CodeExecutionResult>): CodeExecutionResult {
  return { stdout: '', stderr: '', exitCode: 0, ...over };
}

describe('ExecutionResultTable', () => {
  it('result が null なら何も描画しない', () => {
    const { container } = render(
      <ExecutionResultTable result={null} expected={expected} submitError={null} />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it('submitError があればエラーメッセージを表示する', () => {
    render(
      <ExecutionResultTable result={null} expected={expected} submitError="実行に失敗しました" />,
    );
    expect(screen.getByText('実行に失敗しました')).toBeInTheDocument();
  });

  it('exit 0 で出力が期待と一致したときだけ緑の「実行成功・期待する出力と一致」を表示する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":1}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    const status = screen.getByText('実行成功・期待する出力と一致');
    expect(status).toBeInTheDocument();
    expect(status).toHaveClass('text-green-400');
  });

  it('exit 0 でも出力が期待と不一致なら緑にせず、琥珀色の断定形「期待する出力と不一致」を表示する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":2}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    const status = screen.getByText(/実行成功（エラーなし）・期待する出力と不一致/);
    expect(status).toBeInTheDocument();
    expect(status).toHaveClass('text-amber-500');
    expect(status).not.toHaveClass('text-green-400');
  });

  it('不一致のとき、最初に異なる行の行番号と両方の行を表示する', () => {
    render(
      <ExecutionResultTable
        result={result({
          stdout: '✓ testIsEven\n✓ testFindUserMissing\n\nOK (2 tests, 0 assertions)\n',
          exitCode: 0,
        })}
        expected={'✓ testIsEven\n✓ testFindUserMissing\n\nOK (2 tests, 3 assertions)'}
        submitError={null}
      />,
    );
    const hint = screen.getByText(/4 行目が異なります/);
    expect(hint).toBeInTheDocument();
    expect(hint.textContent).toContain('OK (2 tests, 0 assertions)');
    expect(hint.textContent).toContain('OK (2 tests, 3 assertions)');
  });

  it('出力の行数が足りないときは「この行がありません」と表示する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: 'Hello, World!', exitCode: 0 })}
        expected={'Hello, World!\nHello, FreStyle!'}
        submitError={null}
      />,
    );
    const hint = screen.getByText(/2 行目が異なります/);
    expect(hint.textContent).toContain('この行がありません');
    expect(hint.textContent).toContain('Hello, FreStyle!');
  });

  it('出力に余分な行があるときも行番号を特定して表示する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: 'Hello, World!\nextra line', exitCode: 0 })}
        expected={'Hello, World!'}
        submitError={null}
      />,
    );
    const hint = screen.getByText(/2 行目が異なります/);
    expect(hint.textContent).toContain('extra line');
    expect(hint.textContent).toContain('この行がありません');
  });

  it('一致しているときは差分ヒントを表示しない', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":1}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.queryByText(/行目が異なります/)).not.toBeInTheDocument();
  });

  it('末尾改行や行末スペースの差だけなら一致として扱う（サーバ採点の正規化と同じ）', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: 'Hello, World!  \nHello, FreStyle!\n', exitCode: 0 })}
        expected={'Hello, World!\nHello, FreStyle!'}
        submitError={null}
      />,
    );
    expect(screen.getByText('実行成功・期待する出力と一致')).toBeInTheDocument();
  });

  it('期待出力が空の演習では比較せず、従来の「実行成功（エラーなし）」を表示する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: 'something', exitCode: 0 })}
        expected=""
        submitError={null}
      />,
    );
    expect(screen.getByText('実行成功（エラーなし）')).toBeInTheDocument();
    expect(screen.queryByText(/まだ一致していません/)).not.toBeInTheDocument();
  });

  it('exit が 0 以外なら「実行エラー（exit N）」を表示する', () => {
    render(
      <ExecutionResultTable
        result={result({ stderr: 'Parse error', exitCode: 255 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.getByText('実行エラー（exit 255）')).toBeInTheDocument();
    expect(screen.getByText('Parse error')).toBeInTheDocument();
  });

  it('stderr は専用の「エラーメッセージ」行に表示され、アウトプット行は stdout のみになる', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: 'partial', stderr: 'boom', exitCode: 1 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.getByText('エラーメッセージ')).toBeInTheDocument();
    const errorCell = screen.getByText('boom');
    const outputCell = screen.getByText('partial');
    expect(errorCell.closest('tr')).not.toBe(outputCell.closest('tr'));
  });

  it('go のコンパイル失敗（stderr に main.go:N）は「コンパイル時エラーメッセージ」ラベルになる', () => {
    render(
      <ExecutionResultTable
        result={result({ stderr: './main.go:7:9: undefined: foo', exitCode: 1 })}
        expected={expected}
        submitError={null}
        language="go"
      />,
    );
    expect(screen.getByText('コンパイル時エラーメッセージ')).toBeInTheDocument();
  });

  it('go 以外の言語のエラーは汎用の「エラーメッセージ」ラベルのまま', () => {
    render(
      <ExecutionResultTable
        result={result({ stderr: './main.js:3\nSyntaxError: bad', exitCode: 1 })}
        expected={expected}
        submitError={null}
        language="javascript"
      />,
    );
    expect(screen.getByText('エラーメッセージ')).toBeInTheDocument();
    expect(screen.queryByText('コンパイル時エラーメッセージ')).not.toBeInTheDocument();
  });

  it('stderr が空ならエラーメッセージ行を出さない', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":1}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.queryByText(/エラーメッセージ/)).not.toBeInTheDocument();
  });

  it('エラーなしでも出力が空なら「まだ出力がありません」ヒントを出す（空っぽで成功＝正解と誤解させない）', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.getByText(/まだ出力がありません/)).toBeInTheDocument();
  });

  it('出力があるときは「まだ出力がありません」ヒントを出さない', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":1}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.queryByText(/まだ出力がありません/)).not.toBeInTheDocument();
  });

  it('注記でプレビュー比較であること・正誤の確定は提出時であることを明記する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":1}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.getByText(/サーバ側採点/)).toBeInTheDocument();
  });
});
