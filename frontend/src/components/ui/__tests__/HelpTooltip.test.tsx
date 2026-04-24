import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import HelpTooltip from '../HelpTooltip';

describe('HelpTooltip', () => {
  it('初期表示ではツールチップ本文は表示されない', () => {
    render(<HelpTooltip>5軸評価の説明</HelpTooltip>);
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('aria-label を既定の「詳細を表示」で描画する', () => {
    render(<HelpTooltip>内容</HelpTooltip>);
    expect(screen.getByRole('button', { name: '詳細を表示' })).toBeInTheDocument();
  });

  it('label prop をアクセシブルネームとして採用する', () => {
    render(<HelpTooltip label="5軸評価について">内容</HelpTooltip>);
    expect(screen.getByRole('button', { name: '5軸評価について' })).toBeInTheDocument();
  });

  it('クリック（mousedown）でツールチップが開き、再クリックで閉じる', () => {
    render(<HelpTooltip>5軸評価の説明</HelpTooltip>);
    const trigger = screen.getByRole('button', { name: '詳細を表示' });

    fireEvent.mouseDown(trigger);
    expect(screen.getByRole('tooltip')).toHaveTextContent('5軸評価の説明');
    expect(trigger).toHaveAttribute('aria-expanded', 'true');

    fireEvent.mouseDown(trigger);
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
  });

  it('mousedown→focus→click のブラウザ実イベント順序でもツールチップが閉じずに開いたまま', () => {
    // ブラウザでボタン未フォーカス時にクリックすると
    // mousedown → focus → click の順に発火する。
    // onMouseDown で preventDefault しているため focus は発火しない想定、
    // かつ click では開閉トグルを行わないため、最終状態は「開」のまま。
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = screen.getByRole('button', { name: '詳細を表示' });

    fireEvent.mouseDown(trigger);
    // preventDefault 済みのため focus/click が後段で走ったとしても既状態を維持
    fireEvent.focus(trigger);
    fireEvent.click(trigger);

    expect(screen.getByRole('tooltip')).toBeInTheDocument();
    expect(trigger).toHaveAttribute('aria-expanded', 'true');
  });

  it('Enter キーでツールチップをトグル開閉できる', () => {
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = screen.getByRole('button', { name: '詳細を表示' });

    fireEvent.keyDown(trigger, { key: 'Enter' });
    expect(screen.getByRole('tooltip')).toBeInTheDocument();

    fireEvent.keyDown(trigger, { key: 'Enter' });
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('Space キーでもツールチップをトグル開閉できる', () => {
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = screen.getByRole('button', { name: '詳細を表示' });

    fireEvent.keyDown(trigger, { key: ' ' });
    expect(screen.getByRole('tooltip')).toBeInTheDocument();
  });

  it('Escape キーでツールチップが閉じる', () => {
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = screen.getByRole('button', { name: '詳細を表示' });

    fireEvent.mouseDown(trigger);
    expect(screen.getByRole('tooltip')).toBeInTheDocument();

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('外側クリックでツールチップが閉じる', () => {
    render(
      <div>
        <HelpTooltip>説明</HelpTooltip>
        <button type="button">外側</button>
      </div>
    );

    fireEvent.mouseDown(screen.getByRole('button', { name: '詳細を表示' }));
    expect(screen.getByRole('tooltip')).toBeInTheDocument();

    fireEvent.mouseDown(screen.getByRole('button', { name: '外側' }));
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });
});
