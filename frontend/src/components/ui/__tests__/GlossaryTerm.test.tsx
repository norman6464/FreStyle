import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import GlossaryTerm from '../GlossaryTerm';

describe('GlossaryTerm', () => {
  it('用語ラベルを表示する', () => {
    render(<GlossaryTerm term="5軸評価" definition="5 つの軸で評価します" />);
    expect(screen.getByText('5軸評価')).toBeInTheDocument();
  });

  it('ヘルプアイコンに用語名を含む aria-label が付く', () => {
    render(<GlossaryTerm term="5軸評価" definition="説明" />);
    expect(screen.getByRole('button', { name: '5軸評価の意味を表示' })).toBeInTheDocument();
  });

  it('ヘルプアイコンをクリック（mousedown）すると定義が表示される', () => {
    render(<GlossaryTerm term="5軸評価" definition="5 つの軸で評価します" />);
    fireEvent.mouseDown(screen.getByRole('button', { name: '5軸評価の意味を表示' }));
    expect(screen.getByRole('tooltip')).toHaveTextContent('5 つの軸で評価します');
  });

  it('definition に ReactNode を受け取れる', () => {
    render(
      <GlossaryTerm
        term="練習モード"
        definition={<span data-testid="definition-node">強調テキスト</span>}
      />
    );
    fireEvent.mouseDown(screen.getByRole('button', { name: '練習モードの意味を表示' }));
    expect(screen.getByTestId('definition-node')).toBeInTheDocument();
  });
});
