import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';

// monaco は jsdom で動かないため最小モックに差し替える。
// 目的は「Ctrl/Cmd+Enter のアクションが onRun を呼ぶ」配線の検証。
const editorMock = {
  addAction: vi.fn(),
  addCommand: vi.fn(),
  onDidChangeModelContent: vi.fn(() => ({ dispose: vi.fn() })),
  onDidContentSizeChange: vi.fn(() => ({ dispose: vi.fn() })),
  getContentHeight: vi.fn(() => 200),
  getValue: vi.fn(() => ''),
  setValue: vi.fn(),
  getModel: vi.fn(() => ({ isDisposed: () => false })),
  layout: vi.fn(),
  updateOptions: vi.fn(),
  dispose: vi.fn(),
};

vi.mock('monaco-editor/esm/vs/editor/editor.worker?worker', () => ({
  default: class {},
}));

vi.mock('monaco-editor', () => ({
  editor: {
    create: vi.fn(() => editorMock),
    setModelLanguage: vi.fn(),
  },
  KeyMod: { CtrlCmd: 2048 },
  KeyCode: { Enter: 3 },
}));

import CodeEditor from '../CodeEditor';

describe('CodeEditor の実行ショートカット', () => {
  beforeEach(() => {
    editorMock.addAction.mockClear();
  });

  it('Ctrl/Cmd+Enter のアクションが登録され、run が onRun を呼ぶ', () => {
    const onRun = vi.fn();
    render(<CodeEditor value="x" onChange={() => {}} onRun={onRun} />);

    // addAction が Cmd/Ctrl+Enter のキーバインドで登録されている。
    expect(editorMock.addAction).toHaveBeenCalledTimes(1);
    const action = editorMock.addAction.mock.calls[0][0];
    expect(action.keybindings).toEqual([2048 | 3]); // CtrlCmd | Enter

    // 登録された run を呼ぶと onRun が発火する（= ショートカットでコード実行できる）。
    action.run();
    expect(onRun).toHaveBeenCalledTimes(1);
  });

  it('onRun 未指定でも run 実行で例外を投げない', () => {
    render(<CodeEditor value="x" onChange={() => {}} />);
    const action = editorMock.addAction.mock.calls[0][0];
    expect(() => action.run()).not.toThrow();
  });
});
