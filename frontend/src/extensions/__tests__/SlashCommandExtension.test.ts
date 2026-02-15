import { describe, it, expect, vi } from 'vitest';
import { SLASH_COMMANDS } from '../../constants/slashCommands';
import { executeCommand } from '../SlashCommandExtension';

function findCommand(action: string) {
  return SLASH_COMMANDS.find(c => c.action === action)!;
}

describe('SlashCommandExtension', () => {
  function createMockEditor() {
    const chain = {
      focus: vi.fn().mockReturnThis(),
      setParagraph: vi.fn().mockReturnThis(),
      setHeading: vi.fn().mockReturnThis(),
      toggleBulletList: vi.fn().mockReturnThis(),
      toggleOrderedList: vi.fn().mockReturnThis(),
      run: vi.fn(),
    };
    return { chain: vi.fn(() => chain), _chain: chain };
  }

  it('paragraphコマンドでsetParagraphが呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, findCommand('paragraph'));
    expect(editor._chain.setParagraph).toHaveBeenCalled();
    expect(editor._chain.run).toHaveBeenCalled();
  });

  it('heading1コマンドでsetHeading level:1が呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, findCommand('heading1'));
    expect(editor._chain.setHeading).toHaveBeenCalledWith({ level: 1 });
  });

  it('heading2コマンドでsetHeading level:2が呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, findCommand('heading2'));
    expect(editor._chain.setHeading).toHaveBeenCalledWith({ level: 2 });
  });

  it('heading3コマンドでsetHeading level:3が呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, findCommand('heading3'));
    expect(editor._chain.setHeading).toHaveBeenCalledWith({ level: 3 });
  });

  it('bulletListコマンドでtoggleBulletListが呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, findCommand('bulletList'));
    expect(editor._chain.toggleBulletList).toHaveBeenCalled();
  });

  it('orderedListコマンドでtoggleOrderedListが呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, findCommand('orderedList'));
    expect(editor._chain.toggleOrderedList).toHaveBeenCalled();
  });

  it('toggleListコマンドでsetToggleListが呼ばれる', () => {
    const setToggleList = vi.fn();
    const editor = { chain: vi.fn(() => ({ focus: vi.fn().mockReturnThis(), run: vi.fn() })), commands: { setToggleList } };
    executeCommand(editor as never, findCommand('toggleList'));
    expect(setToggleList).toHaveBeenCalled();
  });

  it('imageコマンドでstorage.onImageUploadが呼ばれる', () => {
    const onImageUpload = vi.fn();
    const editor = { ...createMockEditor(), storage: { slashCommand: { onImageUpload } } };
    executeCommand(editor as never, findCommand('image'));
    expect(onImageUpload).toHaveBeenCalled();
  });

  it('imageコマンドでonImageUploadが未設定でもエラーにならない', () => {
    const editor = { ...createMockEditor(), storage: { slashCommand: { onImageUpload: null } } };
    expect(() => executeCommand(editor as never, findCommand('image'))).not.toThrow();
  });

  it('taskListコマンドでtoggleTaskListが呼ばれる', () => {
    const chain = {
      focus: vi.fn().mockReturnThis(),
      toggleTaskList: vi.fn().mockReturnThis(),
      run: vi.fn(),
    };
    const editor = { chain: vi.fn(() => chain) };
    executeCommand(editor as never, findCommand('taskList'));
    expect(chain.toggleTaskList).toHaveBeenCalled();
    expect(chain.run).toHaveBeenCalled();
  });

  it('codeBlockコマンドでtoggleCodeBlockが呼ばれる', () => {
    const chain = {
      focus: vi.fn().mockReturnThis(),
      toggleCodeBlock: vi.fn().mockReturnThis(),
      run: vi.fn(),
    };
    const editor = { chain: vi.fn(() => chain) };
    executeCommand(editor as never, findCommand('codeBlock'));
    expect(chain.toggleCodeBlock).toHaveBeenCalled();
    expect(chain.run).toHaveBeenCalled();
  });

  it('blockquoteコマンドでtoggleBlockquoteが呼ばれる', () => {
    const chain = {
      focus: vi.fn().mockReturnThis(),
      toggleBlockquote: vi.fn().mockReturnThis(),
      run: vi.fn(),
    };
    const editor = { chain: vi.fn(() => chain) };
    executeCommand(editor as never, findCommand('blockquote'));
    expect(chain.toggleBlockquote).toHaveBeenCalled();
    expect(chain.run).toHaveBeenCalled();
  });

  it('horizontalRuleコマンドでsetHorizontalRuleが呼ばれる', () => {
    const chain = {
      focus: vi.fn().mockReturnThis(),
      setHorizontalRule: vi.fn().mockReturnThis(),
      run: vi.fn(),
    };
    const editor = { chain: vi.fn(() => chain) };
    executeCommand(editor as never, findCommand('horizontalRule'));
    expect(chain.setHorizontalRule).toHaveBeenCalled();
    expect(chain.run).toHaveBeenCalled();
  });

  it('tableコマンドでinsertTableが呼ばれる', () => {
    const chain = {
      focus: vi.fn().mockReturnThis(),
      insertTable: vi.fn().mockReturnThis(),
      run: vi.fn(),
    };
    const editor = { chain: vi.fn(() => chain) };
    executeCommand(editor as never, findCommand('table'));
    expect(chain.insertTable).toHaveBeenCalledWith({ rows: 3, cols: 3, withHeaderRow: true });
    expect(chain.run).toHaveBeenCalled();
  });

  it('情報コールアウトでsetCalloutWithTypeが呼ばれる', () => {
    const setCalloutWithType = vi.fn();
    const editor = { chain: vi.fn(() => ({ focus: vi.fn().mockReturnThis(), run: vi.fn() })), commands: { setCalloutWithType } };
    const infoCallout = SLASH_COMMANDS.find(c => c.action === 'callout' && c.attrs?.calloutType === 'info')!;
    executeCommand(editor as never, infoCallout);
    expect(setCalloutWithType).toHaveBeenCalledWith('info', '💡');
  });

  it('警告コールアウトでsetCalloutWithTypeが呼ばれる', () => {
    const setCalloutWithType = vi.fn();
    const editor = { chain: vi.fn(() => ({ focus: vi.fn().mockReturnThis(), run: vi.fn() })), commands: { setCalloutWithType } };
    const warningCallout = SLASH_COMMANDS.find(c => c.action === 'callout' && c.attrs?.calloutType === 'warning')!;
    executeCommand(editor as never, warningCallout);
    expect(setCalloutWithType).toHaveBeenCalledWith('warning', '⚠️');
  });

  it('SLASH_COMMANDSのフィルタリングが正しく動作する', () => {
    const filtered = SLASH_COMMANDS.filter(item =>
      item.label.toLowerCase().includes('見出し')
    );
    expect(filtered).toHaveLength(3);
    expect(filtered.map(f => f.action)).toEqual(['heading1', 'heading2', 'heading3']);
  });
});
