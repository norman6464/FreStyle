import { describe, it, expect } from 'vitest';
import { SLASH_COMMANDS } from '../slashCommands';

describe('SLASH_COMMANDS', () => {
  it('18のコマンドが定義されている', () => {
    expect(SLASH_COMMANDS).toHaveLength(18);
  });

  it('各コマンドに必要なプロパティがある', () => {
    for (const cmd of SLASH_COMMANDS) {
      expect(cmd.label).toBeTruthy();
      expect(cmd.description).toBeTruthy();
      expect(cmd.icon).toBeTruthy();
      expect(cmd.action).toBeTruthy();
    }
  });

  it('全ラベルがユニーク', () => {
    const labels = SLASH_COMMANDS.map(c => c.label);
    expect(new Set(labels).size).toBe(labels.length);
  });

  it('コールアウトが4種類ある', () => {
    const callouts = SLASH_COMMANDS.filter(c => c.action === 'callout');
    expect(callouts).toHaveLength(4);
    expect(callouts.map(c => c.attrs?.calloutType)).toEqual(['info', 'warning', 'success', 'error']);
  });
});
