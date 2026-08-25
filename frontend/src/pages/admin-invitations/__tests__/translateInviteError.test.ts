import { describe, it, expect } from 'vitest';
import { translateInviteError } from '../lib/translateInviteError';

describe('translateInviteError', () => {
  // 6 分岐すべてと、そこに当たらなかったときのフォールバックを網羅する。
  it.each([
    ['super_admin_can_only_invite_company_admin', '運営は会社管理者のみ招待できます。受講者の招待は会社管理者から行ってください。'],
    ['company_admin_can_only_invite_trainee', '会社管理者が招待できるのは受講者のみです。'],
    ['UsernameExistsException', 'このメールアドレスはすでに登録済みです。再招待は不要です。'],
    ['User account already exists', 'このメールアドレスはすでに登録済みです。再招待は不要です。'],
    ['InvalidParameterException', '入力値が不正です。メールアドレス形式を確認してください。'],
    ['LimitExceededException', '招待リクエストが多すぎます。しばらく待ってから再試行してください。'],
    ['TooManyRequestsException', '招待リクエストが多すぎます。しばらく待ってから再試行してください。'],
    ['AccessDeniedException', '権限エラー: バックエンドの IAM ロール設定を確認してください。'],
    ['not authorized', '権限エラー: バックエンドの IAM ロール設定を確認してください。'],
  ])('%s を日本語メッセージに置き換える', (raw, expected) => {
    expect(translateInviteError(raw)).toBe(expected);
  });

  it('エラーコードが本文に埋め込まれていても部分一致で拾う', () => {
    expect(translateInviteError('invite failed: UsernameExistsException (user pool ap-northeast-1_x)')).toBe(
      'このメールアドレスはすでに登録済みです。再招待は不要です。',
    );
  });

  it('先に一致した分岐を優先する（SoD の分岐は Cognito の例外名より先）', () => {
    expect(
      translateInviteError('super_admin_can_only_invite_company_admin / InvalidParameterException'),
    ).toBe('運営は会社管理者のみ招待できます。受講者の招待は会社管理者から行ってください。');
  });

  it('未知のコードは加工せずそのまま返す（backend で日本語化済みの message をそのまま通すため）', () => {
    expect(translateInviteError('招待の作成に失敗しました')).toBe('招待の作成に失敗しました');
    expect(translateInviteError('SomeUnknownException')).toBe('SomeUnknownException');
    expect(translateInviteError('')).toBe('');
  });
});
