import { describe, it, expect } from 'vitest';
import { UI_TIMINGS, UI_DIMENSIONS } from '../uiTimings';

describe('UI_TIMINGS', () => {
  it('コピーフィードバック時間が定義されている', () => {
    expect(UI_TIMINGS.COPY_FEEDBACK_DURATION).toBe(2000);
  });

  it('ツールバー非表示遅延が定義されている', () => {
    expect(UI_TIMINGS.TOOLBAR_HIDE_DELAY).toBe(150);
  });

  it('ブロック挿入ボタン非表示遅延が定義されている', () => {
    expect(UI_TIMINGS.INSERTER_HIDE_DELAY).toBe(200);
  });

  it('マウス移動スロットルが定義されている', () => {
    expect(UI_TIMINGS.MOUSE_MOVE_THROTTLE).toBe(50);
  });
});

describe('UI_DIMENSIONS', () => {
  it('選択ツールバーオフセットが定義されている', () => {
    expect(UI_DIMENSIONS.SELECTION_TOOLBAR_OFFSET_TOP).toBe(48);
  });
});
