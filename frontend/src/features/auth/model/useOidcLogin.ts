import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { buildAuthorizeUrl } from '../lib/oidcAuthUrl';
import { readAuthConfig } from '../lib/authConfig';
import { classifyApiError } from '@/shared/lib/classifyApiError';
import { logger } from '@/shared/lib/logger';

/**
 * ログインを始める操作。
 *
 * 戻り値を合併にしてあるのが要点で、**`available: false` の枝には `start` が無い**。
 * だから「押せるボタンを描いて、押しても何も起きない」というコードは型として
 * 書けない。設定が欠けているとき、呼ぶ側は必ず `available` を見る必要がある。
 *
 * 検査を実行時のガードに置くと、ガードを呼び忘れた経路には効かない。
 * ここは型なので、これから書かれる導線にも毎回効く。
 */
export type OidcLogin =
  | {
      readonly available: true;
      readonly loading: boolean;
      readonly errorMessage: string | null;
      readonly start: (provider?: string, screenHint?: 'signup' | 'signin') => void;
    }
  | {
      readonly available: false;
      /** 欠けている設定の名前。検査と記録のために持つ（人が読む文には出さない）。 */
      readonly missing: readonly string[];
    };

export function useOidcLogin(): OidcLogin {
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // 設定はビルド時に確定しているので、描画のたびに読み直す意味は無い。
  const config = useMemo(() => readAuthConfig(), []);

  // 二重押しの見張り。state と検証値は呼ぶたびに作り直して sessionStorage を
  // 上書きするので、同時に 2 回走らせると、先に飛んだ方の戻りで state が合わなくなる
  // （利用者から見ると「押しただけでログインに失敗する」）。
  // setLoading は次の描画までしか効かないので、描画に依らない印で止める。
  const starting = useRef(false);

  const start = useCallback(
    async (provider?: string, screenHint?: 'signup' | 'signin') => {
      if (config.status !== 'configured') return;
      if (starting.current) return;
      starting.current = true;
      setLoading(true);
      setErrorMessage(null);
      try {
        // 認可 URL の組み立てには要約（code_challenge）の計算が入るので待つ。
        // 待たずに遷移すると、検証値を置く前にページが消えて次に進めなくなる。
        window.location.href = await buildAuthorizeUrl(config, provider, screenHint);
      } catch (err) {
        starting.current = false;
        setLoading(false);
        setErrorMessage(classifyApiError(err, 'ログイン画面へ移動できませんでした。'));
      }
    },
    [config],
  );

  // 描画のたびに出さない。設定はビルド時に確定しているので、1 回出れば足りる。
  useEffect(() => {
    if (config.status === 'unconfigured') {
      logger.error('認可の設定が揃っていないため、ログインを開始できません', {
        missing: config.missing,
      });
    }
  }, [config]);

  if (config.status !== 'configured') {
    return { available: false, missing: config.missing };
  }

  return { available: true, loading, errorMessage, start };
}
