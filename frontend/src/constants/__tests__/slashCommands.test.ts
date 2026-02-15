import { describe, it, expect } from 'vitest';
import { SLASH_COMMANDS } from '../slashCommands';

describe('SLASH_COMMANDS', () => {
  it('15のコマンドが定義されている', () => {
    expect(SLASH_COMMANDS).toHaveLength(15);
  });

  it('各コマンドに必要なプロパティがある', () => {
    for (const cmd of SLASH_COMMANDS) {
      expect(cmd.label).toBeTruthy();
      expect(cmd.description).toBeTruthy();
      expect(cmd.icon).toBeTruthy();
      expect(cmd.action).toBeTruthy();
    }
  });

  it('全アクションがユニーク', () => {
    const actions = SLASH_COMMANDS.map(c => c.action);
    expect(new Set(actions).size).toBe(actions.length);
  });
});
