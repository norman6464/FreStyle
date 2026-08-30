import { useCallback, useEffect, useState } from 'react';

import { AdminInvitationRepository, AdminInvitation, CreateInvitationForm } from '@/entities/invitation';
import type { InvitationMethod } from '@/entities/invitation';
import { CompanyRepository, Company } from '@/entities/company';
import { AuthRepository, UserInfo } from '@/entities/user';
import { logger } from '@/shared/lib/logger';

import { extractApiErrorMessage } from '../lib/extractApiErrorMessage';
import { translateInviteError } from '../lib/translateInviteError';

const EMPTY_FORM: CreateInvitationForm = {
  companyId: 0,
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
 * 招待一覧・会社一覧の取得、招待の作成（招待リンク / 初期パスワード）、招待の取り消しを扱う。
 * 通過条件（管理者かどうか）はルート側の RequireRole が持つのでここでは判定しない。
 */
export function useAdminInvitations() {
  const [me, setMe] = useState<UserInfo | null>(null);
  const [invitations, setInvitations] = useState<AdminInvitation[]>([]);
  const [companies, setCompanies] = useState<Company[]>([]);
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

  // 認可境界（SoD）に応じて招待 UI を切り替えるため、自分の role / companyId を取得する。
  // backend 側でも同じ境界を強制しているので、フロントは UX 改善目的（不可能な選択肢を見せない）。
  //   - super_admin → role=company_admin で固定 / company は任意選択
  //   - company_admin → role=trainee で固定 / company は自社固定
  const isSuperAdmin = me?.role === 'super_admin';
  const isCompanyAdmin = me?.role === 'company_admin';

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      // 会社一覧 API は super_admin 専用。company_admin で呼ぶと 403 になり、
      // Promise.all ごと reject して招待一覧まで巻き込むため、先に自分の role を
      // 確認してから super_admin のときだけ取得する（company_admin は自社固定で不要）。
      const user = await AuthRepository.getCurrentUser();
      const [invitationList, companyList] = await Promise.all([
        AdminInvitationRepository.list(),
        user.role === 'super_admin' ? CompanyRepository.list() : Promise.resolve<Company[]>([]),
      ]);
      setMe(user);
      setInvitations(invitationList);
      setCompanies(companyList);

      // 役割に応じてフォームの初期値を上書きする。
      // company_admin は自社固定（InvitationForm が company_admin 用の選択 UI 自体を出さない）ため
      // ここでは super_admin 向けの既定値（会社一覧の先頭）だけを算出する。
      const defaultRole: CreateInvitationForm['role'] =
        user.role === 'super_admin' ? 'company_admin' : 'trainee';
      const defaultCompanyId = companyList[0]?.id ?? 0;
      setForm((f) => ({
        ...f,
        role: defaultRole,
        companyId: f.companyId === 0 ? defaultCompanyId : f.companyId,
      }));
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
    // 会社の選択が要るのは super_admin だけ。company_admin の招待先は
    // backend が actor 自身の所属ワークスペースに固定するため、フォームでは選ばせない。
    if (isSuperAdmin && !form.companyId) {
      setError('会社を選択してください');
      return;
    }
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
      setForm((f) => ({ ...EMPTY_FORM, companyId: f.companyId }));
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
    companies,
    loading,
    error,
    success,
    form,
    setForm,
    submitting,
    submit,
    method,
    setMethod,
    isSuperAdmin,
    isCompanyAdmin,
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
