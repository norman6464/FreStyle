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
    expect(status.className).toContain('text-green-400');
  });

  it('exit 0 でも出力が期待と不一致なら緑にせず、琥珀色の「まだ一致していません」を表示する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":2}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    const status = screen.getByText(/実行成功（エラーなし）・期待する出力とはまだ一致していません/);
    expect(status).toBeInTheDocument();
    expect(status.className).toContain('text-amber-500');
    expect(status.className).not.toContain('text-green-400');
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
