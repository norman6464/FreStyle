import { describe, it, expect, beforeEach } from 'vitest';
import {
  saveInvitationToken,
  readInvitationToken,
  clearInvitationToken,
  consumeInvitationToken,
} from '../invitationToken';

describe('invitationToken', () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it('save / read で同じ値が取り出せる', () => {
    saveInvitationToken('abc-123');
    expect(readInvitationToken()).toBe('abc-123');
  });

  it('空文字は保存しない（誤って空文字を上書き保存しない）', () => {
    saveInvitationToken('first');
    saveInvitationToken('');
    expect(readInvitationToken()).toBe('first');
  });

  it('clear で削除される', () => {
    saveInvitationToken('abc');
    clearInvitationToken();
    expect(readInvitationToken()).toBeNull();
  });

  it('consume は read + clear を 1 回で行う（再利用防止）', () => {
    saveInvitationToken('one-shot');
    expect(consumeInvitationToken()).toBe('one-shot');
    expect(readInvitationToken()).toBeNull();
    expect(consumeInvitationToken()).toBeNull();
  });
});
