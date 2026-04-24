import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import HelpTooltip from '../HelpTooltip';

/**
 * HelpTooltip は Disclosure パターンを採用しているため、
 * パネルの存在は aria-expanded の状態 と aria-controls が指す
 * パネルID の有無で検証する（role="tooltip" は使わない）。
 */
function getTrigger() {
  return screen.getByRole('button', { name: '詳細を表示' });
}

function getPanel(trigger: HTMLElement) {
  const panelId = trigger.getAttribute('aria-controls');
  if (!panelId) throw new Error('aria-controls が未設定');
  return document.getElementById(panelId);
}

describe('HelpTooltip', () => {
  it('初期表示ではパネルは表示されない', () => {
    render(<HelpTooltip>5軸評価の説明</HelpTooltip>);
    const trigger = getTrigger();
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
    expect(getPanel(trigger)).toBeNull();
  });

  it('aria-label を既定の「詳細を表示」で描画する', () => {
    render(<HelpTooltip>内容</HelpTooltip>);
    expect(screen.getByRole('button', { name: '詳細を表示' })).toBeInTheDocument();
  });

  it('label prop をアクセシブルネームとして採用する', () => {
    render(<HelpTooltip label="5軸評価について">内容</HelpTooltip>);
    expect(screen.getByRole('button', { name: '5軸評価について' })).toBeInTheDocument();
  });

  it('aria-controls が常にパネルIDを指す（Disclosure パターン）', () => {
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = getTrigger();
    expect(trigger).toHaveAttribute('aria-controls');
    expect(trigger.getAttribute('aria-controls')).not.toBe('');
  });

  it('クリック（mousedown）でパネルが開き、再クリックで閉じる', () => {
    render(<HelpTooltip>5軸評価の説明</HelpTooltip>);
    const trigger = getTrigger();

    fireEvent.mouseDown(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'true');
    expect(getPanel(trigger)).toHaveTextContent('5軸評価の説明');

    fireEvent.mouseDown(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
    expect(getPanel(trigger)).toBeNull();
  });

  it('mousedown→focus→click のブラウザ実イベント順序でもパネルが閉じずに開いたまま', () => {
    // ブラウザでボタン未フォーカス時にクリックすると
    // mousedown → focus → click の順に発火する。
    // onMouseDown で preventDefault しているため focus は発火しない想定、
    // かつ click では開閉トグルを行わないため、最終状態は「開」のまま。
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = getTrigger();

    fireEvent.mouseDown(trigger);
    fireEvent.focus(trigger);
    fireEvent.click(trigger);

    expect(trigger).toHaveAttribute('aria-expanded', 'true');
    expect(getPanel(trigger)).toBeInTheDocument();
  });

  it('Enter キーでパネルをトグル開閉できる', () => {
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = getTrigger();

    fireEvent.keyDown(trigger, { key: 'Enter' });
    expect(trigger).toHaveAttribute('aria-expanded', 'true');

    fireEvent.keyDown(trigger, { key: 'Enter' });
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
  });

  it('Space キーでもパネルをトグル開閉できる', () => {
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = getTrigger();

    fireEvent.keyDown(trigger, { key: ' ' });
    expect(trigger).toHaveAttribute('aria-expanded', 'true');

    fireEvent.keyDown(trigger, { key: ' ' });
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
  });

  it('Escape キーでパネルが閉じる', () => {
    render(<HelpTooltip>説明</HelpTooltip>);
    const trigger = getTrigger();

    fireEvent.mouseDown(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
  });

  it('外側クリックでパネルが閉じる', () => {
    render(
      <div>
        <HelpTooltip>説明</HelpTooltip>
        <button type="button">外側</button>
      </div>
    );
    const trigger = getTrigger();

    fireEvent.mouseDown(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'true');

    fireEvent.mouseDown(screen.getByRole('button', { name: '外側' }));
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
  });

  it('wrapper 外にフォーカスが移ったらパネルが閉じる（onBlur 経由）', () => {
    render(
      <div>
        <HelpTooltip>説明</HelpTooltip>
        <button type="button" data-testid="外側ボタン">
          外側
        </button>
      </div>
    );
    const trigger = getTrigger();
    // フォーカス→開く（React のシンセティックイベントを fireEvent 経由で発火）
    fireEvent.focus(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'true');
    // relatedTarget が wrapper 外に飛ぶ blur を発火
    const outsideButton = screen.getByTestId('外側ボタン');
    fireEvent.blur(trigger, { relatedTarget: outsideButton });
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
  });
});
