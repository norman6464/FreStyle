interface ApiErrorLike {
  response?: { data?: { message?: string; error?: string } };
}

/** API エラーの本文（message → error の順）を取り出す。取れないときは fallback を返す。 */
export function extractApiErrorMessage(err: unknown, fallback: string): string {
  const data = (err as ApiErrorLike)?.response?.data;
  return data?.message || data?.error || fallback;
}
