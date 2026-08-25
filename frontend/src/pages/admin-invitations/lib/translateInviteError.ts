/**
 * translateInviteError — 招待 API のエラー本文を日本語に置き換える。
 *
 * バックエンドが英語のエラーコード（SoD 違反コード / Cognito の例外名）をそのまま返した場合の
 * フォールバック日本語化。バックエンドでカテゴリ化済みの日本語メッセージ（message フィールド）は
 * どの分岐にも当たらないのでそのまま通る。
 */
export function translateInviteError(raw: string): string {
  if (raw.includes('super_admin_can_only_invite_company_admin')) {
    return '運営は会社管理者のみ招待できます。受講者の招待は会社管理者から行ってください。';
  }
  if (raw.includes('company_admin_can_only_invite_trainee')) {
    return '会社管理者が招待できるのは受講者のみです。';
  }
  if (raw.includes('UsernameExistsException') || raw.includes('User account already exists')) {
    return 'このメールアドレスはすでに登録済みです。再招待は不要です。';
  }
  if (raw.includes('InvalidParameterException')) {
    return '入力値が不正です。メールアドレス形式を確認してください。';
  }
  if (raw.includes('LimitExceededException') || raw.includes('TooManyRequestsException')) {
    return '招待リクエストが多すぎます。しばらく待ってから再試行してください。';
  }
  if (raw.includes('AccessDeniedException') || raw.includes('not authorized')) {
    return '権限エラー: バックエンドの IAM ロール設定を確認してください。';
  }
  return raw;
}
