import { ChangeEvent, FormEvent, useState } from 'react';
import {
  CompanyApplicationForm,
  CompanyApplicationRepository,
} from '../repositories/CompanyApplicationRepository';
import { getApiError } from '@/shared/lib/classifyApiError';

const GENERIC_ERROR = '送信に失敗しました。時間をおいて再度お試しください。';

// serverErrorMessage は backend の {message} / {error} を取り出す（無ければ汎用文言）。
function serverErrorMessage(err: unknown): string {
  const { serverCode, serverMessage } = getApiError(err);
  return serverMessage || serverCode || GENERIC_ERROR;
}

const EMPTY: CompanyApplicationForm = {
  companyName: '',
  applicantName: '',
  email: '',
  message: '',
};

/** useCompanyApplication は企業利用申請フォームの状態と送信を扱う。 */
export function useCompanyApplication() {
  const [form, setForm] = useState<CompanyApplicationForm>(EMPTY);
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleChange = (
    e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!form.companyName.trim() || !form.applicantName.trim() || !form.email.trim()) {
      setError('会社名・お名前・メールアドレスは必須です。');
      return;
    }
    setLoading(true);
    try {
      await CompanyApplicationRepository.apply(form);
      setSubmitted(true);
      setForm(EMPTY);
    } catch (err) {
      setError(serverErrorMessage(err));
    } finally {
      setLoading(false);
    }
  };

  return { form, loading, submitted, error, handleChange, handleSubmit };
}
