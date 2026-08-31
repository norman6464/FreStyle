import { useCallback, useEffect, useState } from 'react';

import { AdminInvitationRepository, AdminInvitation, CreateInvitationForm } from '@/entities/invitation';
import type { InvitationMethod } from '@/entities/invitation';
import { logger } from '@/shared/lib/logger';

import { extractApiErrorMessage } from '../lib/extractApiErrorMessage';
import { translateInviteError } from '../lib/translateInviteError';

const EMPTY_FORM: CreateInvitationForm = {
  email: '',
  role: 'trainee',
  displayName: '',
};

/** 初期パスワード方式で 1 度だけ表示する発行結果。閉じると再取得できない。 */
export interface IssuedPassword {
  email: string;
  password: string;
}

/**
 * useAdminInvitations — メンバー招待ページの状態管理フック。
 * 招待一覧の取得、招待の作成（招待リンク / 初期パスワード）、招待の取り消しを扱う。
 * 通過条件（管理者かどうか）はルート側が持ち、越権はどのみち backend が弾くので判定しない。
 */
export function useAdminInvitations() {
  const [invitations, setInvitations] = useState<AdminInvitation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<CreateInvitationForm>(EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);
  const [cancelTarget, setCancelTarget] = useState<AdminInvitation | null>(null);
  const [canceling, setCanceling] = useState(false);
  const [method, setMethod] = useState<InvitationMethod>('magic_link');
  const [issuedPassword, setIssuedPassword] = useState<IssuedPassword | null>(null);
  const [copied, setCopied] = useState(false);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      // 招待先も役職も固定なので、この画面は自分が誰かを見なくてよい。
      setInvitations(await AdminInvitationRepository.list());
      setError(null);
    } catch (e) {
      setError('データの取得に失敗しました');
      logger.error(e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const submit = async () => {
    setSubmitting(true);
    setError(null);
    setSuccess(null);
    try {
      if (method === 'temporary_password') {
        const { invitation, temporaryPassword } =
          await AdminInvitationRepository.createWithTemporaryPassword(form);
        // 一時パスワードは 1 度だけ表示する。保存も再取得もできない。
        setIssuedPassword({ email: invitation.email, password: temporaryPassword });
        setCopied(false);
        setSuccess(null);
      } else {
        const created = await AdminInvitationRepository.create(form);
        setSuccess(
          `${created.email} 宛に招待メールを送信しました。受信者にメール内のリンクを開いてもらい、画面の案内に従ってログインしてもらってください。`
        );
      }
      setForm(EMPTY_FORM);
      await fetchAll();
    } catch (err: unknown) {
      const raw = extractApiErrorMessage(err, '招待の作成に失敗しました');
      setError(translateInviteError(raw));
      logger.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const requestCancel = (inv: AdminInvitation) => {
    setError(null);
    setCancelTarget(inv);
  };

  const closeCancelModal = () => {
    if (canceling) return;
    setCancelTarget(null);
  };

  const confirmCancel = async () => {
    if (!cancelTarget) return;
    setCanceling(true);
    try {
      await AdminInvitationRepository.cancel(cancelTarget.id);
      setCancelTarget(null);
      await fetchAll();
    } catch (err) {
      setError('招待のキャンセルに失敗しました');
      logger.error(err);
    } finally {
      setCanceling(false);
    }
  };

  // 発行済みの初期パスワードをクリップボードへ写す。失敗しても画面は壊さず表示だけ戻す。
  const copyIssuedPassword = async () => {
    if (!issuedPassword) return;
    try {
      await navigator.clipboard.writeText(issuedPassword.password);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  };

  const dismissIssuedPassword = () => setIssuedPassword(null);

  return {
    invitations,
    loading,
    error,
    success,
    form,
    setForm,
    submitting,
    submit,
    method,
    setMethod,
    cancelTarget,
    canceling,
    requestCancel,
    closeCancelModal,
    confirmCancel,
    issuedPassword,
    copied,
    copyIssuedPassword,
    dismissIssuedPassword,
  };
}
