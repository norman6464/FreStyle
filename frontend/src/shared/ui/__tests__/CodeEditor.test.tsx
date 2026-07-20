import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';

// monaco は jsdom で動かないため最小モックに差し替える。
// 目的は「Ctrl/Cmd+Enter のアクションが onRun を呼ぶ」配線の検証。
const modelMock = {
  isDisposed: () => false,
  getLineCount: vi.fn(() => 10),
  getLineMaxColumn: vi.fn(() => 20),
};

const decorationsMock = { clear: vi.fn(), set: vi.fn() };

const editorMock = {
  addAction: vi.fn(),
  addCommand: vi.fn(),
  onDidChangeModelContent: vi.fn(() => ({ dispose: vi.fn() })),
  onDidContentSizeChange: vi.fn(() => ({ dispose: vi.fn() })),
  getContentHeight: vi.fn(() => 200),
  getValue: vi.fn(() => ''),
  setValue: vi.fn(),
  getModel: vi.fn(() => modelMock),
  layout: vi.fn(),
  updateOptions: vi.fn(),
  createDecorationsCollection: vi.fn(() => decorationsMock),
  dispose: vi.fn(),
};

const setModelMarkersMock = vi.fn();

vi.mock('monaco-editor/esm/vs/editor/editor.worker?worker', () => ({
  default: class {},
}));

vi.mock('monaco-editor', () => ({
  editor: {
    create: vi.fn(() => editorMock),
    setModelLanguage: vi.fn(),
    setModelMarkers: (...args: unknown[]) => setModelMarkersMock(...args),
  },
  MarkerSeverity: { Error: 8 },
  Range: class {
    constructor(
      public startLineNumber: number,
      public startColumn: number,
      public endLineNumber: number,
      public endColumn: number,
    ) {}
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

describe('CodeEditor のエラー行マーカー (FRESTYLE-117)', () => {
  beforeEach(() => {
    setModelMarkersMock.mockClear();
    editorMock.createDecorationsCollection.mockClear();
    editorMock.updateOptions.mockClear();
  });

  it('errorMarkers を渡すと monaco マーカーとガター装飾が設定され、glyphMargin が有効になる', () => {
    render(
      <CodeEditor
        value="x"
        onChange={() => {}}
        errorMarkers={[{ line: 3, message: 'undefined: foo' }]}
      />,
    );
    expect(setModelMarkersMock).toHaveBeenCalled();
    const markers = setModelMarkersMock.mock.calls.at(-1)![2];
    expect(markers).toHaveLength(1);
    expect(markers[0].startLineNumber).toBe(3);
    expect(markers[0].message).toBe('undefined: foo');
    expect(editorMock.createDecorationsCollection).toHaveBeenCalledTimes(1);
    expect(editorMock.updateOptions).toHaveBeenCalledWith({ glyphMargin: true });
  });

  it('行番号は 1〜行数の範囲にクランプされる', () => {
    render(
      <CodeEditor
        value="x"
        onChange={() => {}}
        errorMarkers={[{ line: 999, message: 'too far' }]}
      />,
    );
    const markers = setModelMarkersMock.mock.calls.at(-1)![2];
    expect(markers[0].startLineNumber).toBe(10); // modelMock.getLineCount() = 10
  });

  it('errorMarkers が空ならマーカーは空で glyphMargin も無効', () => {
    render(<CodeEditor value="x" onChange={() => {}} errorMarkers={[]} />);
    const markers = setModelMarkersMock.mock.calls.at(-1)![2];
    expect(markers).toHaveLength(0);
    expect(editorMock.createDecorationsCollection).not.toHaveBeenCalled();
    expect(editorMock.updateOptions).toHaveBeenCalledWith({ glyphMargin: false });
  });
});
