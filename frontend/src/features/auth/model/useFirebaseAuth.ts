import { useCallback, useMemo, useRef, useState } from 'react';
import {
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signInWithPopup,
  GoogleAuthProvider,
  sendPasswordResetEmail,
} from 'firebase/auth';
import { readFirebaseAuthConfig } from '@/shared/lib/auth/firebaseConfig';
import { getFirebaseAuth } from '@/shared/lib/auth/firebaseApp';
import { classifyFirebaseError } from '@/shared/lib/auth/firebaseErrorMessage';

/**
 * GCIP（Firebase Authentication 互換）でのサインイン・サインアップ・
 * パスワード再設定・Google 連携をまとめて提供するフック。
 *
 * `features/auth/model/useOidcLogin.ts`（ローカル Dex 向け）と同じく、設定が
 * 欠けている枝には操作そのものが存在しない判別可能な合併を返す。押せるのに
 * 何も起きないボタンを型で書けなくする（`AuthUnavailableNotice` と組み合わせる）。
 */
export type FirebaseAuthActions =
  | {
      readonly available: true;
      readonly loading: boolean;
      readonly errorMessage: string | null;
      /** 成功したら true。失敗時は errorMessage が立ち false を返す（例外は投げない）。 */
      readonly signInWithEmail: (email: string, password: string) => Promise<boolean>;
      readonly signUpWithEmail: (email: string, password: string) => Promise<boolean>;
      readonly signInWithGoogle: () => Promise<boolean>;
      readonly sendPasswordReset: (email: string) => Promise<boolean>;
    }
  | {
      readonly available: false;
      readonly missing: readonly string[];
    };

export function useFirebaseAuth(): FirebaseAuthActions {
  const config = useMemo(() => readFirebaseAuthConfig(), []);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  // 多重送信を防ぐ（連打でサインイン要求が複数飛ぶのを防ぐ）。ページ遷移で
  // アンマウントされるまで生き続けてよいので useRef で十分（useOidcLogin と同じ形）。
  const inFlight = useRef(false);

  const run = useCallback(
    async (action: () => Promise<void>, fallback: string): Promise<boolean> => {
      if (config.status !== 'configured') return false;
      if (inFlight.current) return false;
      inFlight.current = true;
      setLoading(true);
      setErrorMessage(null);
      try {
        await action();
        return true;
      } catch (err) {
        setErrorMessage(classifyFirebaseError(err, fallback));
        return false;
      } finally {
        inFlight.current = false;
        setLoading(false);
      }
    },
    [config],
  );

  const signInWithEmail = useCallback(
    (email: string, password: string) =>
      run(async () => {
        const auth = getFirebaseAuth();
        if (!auth) throw new Error('firebase auth not initialized');
        await signInWithEmailAndPassword(auth, email, password);
      }, 'ログインできませんでした。'),
    [run],
  );

  const signUpWithEmail = useCallback(
    (email: string, password: string) =>
      run(async () => {
        const auth = getFirebaseAuth();
        if (!auth) throw new Error('firebase auth not initialized');
        await createUserWithEmailAndPassword(auth, email, password);
      }, 'アカウントを作成できませんでした。'),
    [run],
  );

  const signInWithGoogle = useCallback(
    () =>
      run(async () => {
        const auth = getFirebaseAuth();
        if (!auth) throw new Error('firebase auth not initialized');
        await signInWithPopup(auth, new GoogleAuthProvider());
      }, 'Google でのログインに失敗しました。'),
    [run],
  );

  const sendPasswordReset = useCallback(
    (email: string) =>
      run(async () => {
        const auth = getFirebaseAuth();
        if (!auth) throw new Error('firebase auth not initialized');
        await sendPasswordResetEmail(auth, email);
      }, 'パスワード再設定メールを送信できませんでした。'),
    [run],
  );

  if (config.status !== 'configured') {
    return { available: false, missing: config.missing };
  }

  return {
    available: true,
    loading,
    errorMessage,
    signInWithEmail,
    signUpWithEmail,
    signInWithGoogle,
    sendPasswordReset,
  };
}
