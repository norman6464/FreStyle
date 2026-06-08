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

  it('exit 0 は「実行成功（エラーなし）」と表示し、正誤ではないことを明記する', () => {
    render(
      <ExecutionResultTable
        result={result({ stdout: '{"id":1}', exitCode: 0 })}
        expected={expected}
        submitError={null}
      />,
    );
    expect(screen.getByText('実行成功（エラーなし）')).toBeInTheDocument();
    // 「Success」単独で正誤と誤解させる旧文言は使わない
    expect(screen.queryByText('Success')).not.toBeInTheDocument();
    expect(screen.getByText(/正誤の判定ではありません/)).toBeInTheDocument();
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
});
