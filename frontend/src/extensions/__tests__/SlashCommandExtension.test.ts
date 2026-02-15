import { describe, it, expect, vi } from 'vitest';
import { SLASH_COMMANDS } from '../../constants/slashCommands';
import { executeCommand } from '../SlashCommandExtension';

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
    executeCommand(editor as never, SLASH_COMMANDS[0]);
    expect(editor._chain.setParagraph).toHaveBeenCalled();
    expect(editor._chain.run).toHaveBeenCalled();
  });

  it('heading1コマンドでsetHeading level:1が呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, SLASH_COMMANDS[1]);
    expect(editor._chain.setHeading).toHaveBeenCalledWith({ level: 1 });
  });

  it('heading2コマンドでsetHeading level:2が呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, SLASH_COMMANDS[2]);
    expect(editor._chain.setHeading).toHaveBeenCalledWith({ level: 2 });
  });

  it('heading3コマンドでsetHeading level:3が呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, SLASH_COMMANDS[3]);
    expect(editor._chain.setHeading).toHaveBeenCalledWith({ level: 3 });
  });

  it('bulletListコマンドでtoggleBulletListが呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, SLASH_COMMANDS[4]);
    expect(editor._chain.toggleBulletList).toHaveBeenCalled();
  });

  it('orderedListコマンドでtoggleOrderedListが呼ばれる', () => {
    const editor = createMockEditor();
    executeCommand(editor as never, SLASH_COMMANDS[5]);
    expect(editor._chain.toggleOrderedList).toHaveBeenCalled();
  });

  it('toggleListコマンドでsetToggleListが呼ばれる', () => {
    const setToggleList = vi.fn();
    const editor = { chain: vi.fn(() => ({ focus: vi.fn().mockReturnThis(), run: vi.fn() })), commands: { setToggleList } };
    executeCommand(editor as never, SLASH_COMMANDS[6]);
    expect(setToggleList).toHaveBeenCalled();
  });

  it('SLASH_COMMANDSのフィルタリングが正しく動作する', () => {
    const filtered = SLASH_COMMANDS.filter(item =>
      item.label.toLowerCase().includes('見出し')
    );
    expect(filtered).toHaveLength(3);
    expect(filtered.map(f => f.action)).toEqual(['heading1', 'heading2', 'heading3']);
  });
});
