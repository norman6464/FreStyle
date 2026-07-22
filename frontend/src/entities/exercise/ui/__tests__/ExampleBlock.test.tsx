import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ExampleBlock from '../ExampleBlock';
import type { MasterExerciseExample } from '../../model/types';

const example = (over: Partial<MasterExerciseExample> = {}): MasterExerciseExample => ({
  inputText: '3\n',
  expectedOutput: 'Fizz\n',
  ...over,
} as MasterExerciseExample);

describe('ExampleBlock', () => {
  it('入力と期待する出力を表示する', () => {
    render(<ExampleBlock index={1} total={1} example={example()} />);
    expect(screen.getByText('入力される値')).toBeInTheDocument();
    expect(screen.getByText('期待する出力')).toBeInTheDocument();
  });

  it('例が複数あるときは見出しに番号を付ける', () => {
    render(<ExampleBlock index={2} total={3} example={example()} />);
    expect(screen.getByText('入力される値 2')).toBeInTheDocument();
    expect(screen.getByText('期待する出力 2')).toBeInTheDocument();
  });

  it('入力が空のときは「ありません。」と標準入力の補足を出す', () => {
    render(<ExampleBlock index={1} total={1} example={example({ inputText: '' })} />);
    expect(screen.getByText('ありません。')).toBeInTheDocument();
    expect(screen.getByText(/標準入力から渡されます/)).toBeInTheDocument();
  });

  it('入力があるときは標準入力の補足を出さない', () => {
    render(<ExampleBlock index={1} total={1} example={example()} />);
    expect(screen.queryByText(/標準入力から渡されます/)).not.toBeInTheDocument();
  });
});
