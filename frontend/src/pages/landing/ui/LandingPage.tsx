import { useEffect } from 'react';
import { Link, Navigate } from 'react-router-dom';
import PublicHeader from '@/shared/ui/PublicHeader';
import { useAppSelector } from '@/shared/lib/store';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';

const SITE_URL = 'https://normanblog.com/';

const FEATURES: { title: string; body: string }[] = [
  {
    title: 'コース学習',
    body: 'Git・Docker・Linux・Go・PostgreSQL からクリーンアーキテクチャまで、実務直結のコースを章立てで学べる。進捗はコースごとに記録され、続きから再開できる。',
  },
  {
    title: 'コーディング演習',
    body: 'ブラウザ上のエディタで書いて即採点。PHP・Java・JavaScript・TypeScript・Go・Ruby・C・C++・SQL など多言語に対応し、手を動かして身につける。',
  },
  {
    title: 'AI チャットで質問',
    body: '詰まったところを AI にその場で相談。学習の流れを止めずに、疑問を解消しながら進められる。',
  },
  {
    title: '学習レポート',
    body: '取り組んだ量や進み具合をレポートで可視化。受講者は自分の伸びを、研修担当はチームの状況を把握できる。',
  },
  {
    title: 'メンバーの学習状況（管理者）',
    body: '企業の研修担当は、メンバーごとの進捗・つまずきを一覧で確認。招待もマジックリンクで簡単に行える。',
  },
  {
    title: '迷わない導線',
    body: '「何を学ぶか」で迷わないよう、学習領域→コース→章の順に整理。初学者がひとりでも進められる設計。',
  },
];

const STEPS: { title: string; body: string }[] = [
  { title: '1. 利用申請', body: '企業の研修担当が利用申請を送信。折り返しご案内します。' },
  { title: '2. メンバー招待', body: '受講者をマジックリンクで招待。受け取ったリンクから参加できます。' },
  { title: '3. 学習開始', body: 'コース学習とコーディング演習で、手を動かしながら研修を進めます。' },
  { title: '4. 進捗の確認', body: '学習レポートで、受講者・研修担当の双方が伸びを確認できます。' },
];

const FAQS: { q: string; a: string }[] = [
  {
    q: 'FreStyle とは何ですか？',
    a: 'FreStyle は、新卒 IT エンジニア向けの統合研修プラットフォームです。コース学習・多言語のコーディング演習・AI チャット・学習レポートを1つにまとめ、研修の実施と進捗管理を支援します。',
  },
  {
    q: '誰が使うサービスですか？',
    a: '新卒・若手エンジニアの受講者と、その研修を運営する企業の研修担当者が対象です。受講者は学習を、担当者はメンバーの学習状況の把握を行えます。',
  },
  {
    q: 'どんな言語の演習ができますか？',
    a: 'PHP・Java・JavaScript・TypeScript・Go・Ruby・C・C++・SQL などに対応しています。ブラウザ上で書いてそのまま採点できます。',
  },
  {
    q: '利用を始めるには？',
    a: '企業の研修担当者が利用申請を送るところから始まります。受講者は担当者からの招待リンクで参加します。',
  },
];

/**
 * LandingPage は未ログイン/検索ボット向けの公開トップ。SEO のインデックス対象。
 * API は叩かず純静的に描画し、ビルド時プリレンダー(Playwright)で中身入り HTML を配信する。
 * ログイン済みで来た場合はダッシュボードへ送る（クライアント遷移時のみ。初回ロードは
 * 認証状態未確定なので LP を表示する = プリレンダーも LP になる）。
 */
export default function LandingPage() {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);

  useDocumentMeta({
    title: 'FreStyle | 新卒ITエンジニア向け研修プラットフォーム',
    description:
      'FreStyle は新卒・若手 IT エンジニア向けの研修プラットフォーム。コース学習・多言語のコーディング演習・AI チャット・学習レポートで、研修の実施と進捗管理を支援します。',
    canonical: SITE_URL,
  });

  // プリレンダーが「描画完了」を検知するための目印（マウント後に付与）。
  useEffect(() => {
    document.documentElement.setAttribute('data-prerender-ready', 'true');
    return () => document.documentElement.removeAttribute('data-prerender-ready');
  }, []);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface)] text-[var(--color-text-primary)]">
      <PublicHeader />

      <main>
        {/* ヒーロー */}
        <section className="mx-auto max-w-5xl px-6 pt-16 pb-14 text-center">
          <h1 className="text-3xl sm:text-5xl font-bold leading-tight">
            FreStyle — 新卒ITエンジニア向け研修プラットフォーム
          </h1>
          <p className="mt-6 text-lg sm:text-xl text-[var(--color-text-secondary)] max-w-3xl mx-auto">
            コース学習・多言語のコーディング演習・AI チャット・学習レポートを1つに。
            手を動かして学び、研修の進捗をまとめて管理できます。
          </p>
          <div className="mt-9 flex flex-col sm:flex-row gap-3 justify-center">
            <Link
              to="/company-application"
              className="inline-flex items-center justify-center rounded-lg bg-[var(--color-accent)] px-6 py-3 font-semibold text-white shadow-sm hover:opacity-90 transition"
            >
              企業の方：導入・利用申請
            </Link>
            <Link
              to="/login"
              className="inline-flex items-center justify-center rounded-lg border border-[var(--color-border-hover)] px-6 py-3 font-semibold text-[var(--color-text-primary)] hover:bg-[var(--color-surface-2)] transition"
            >
              受講者の方：ログイン
            </Link>
          </div>
        </section>

        {/* FreStyle とは */}
        <section className="bg-[var(--color-surface-1)] border-y border-[var(--color-surface-3)]">
          <div className="mx-auto max-w-4xl px-6 py-14">
            <h2 className="text-2xl font-bold">FreStyle とは</h2>
            <p className="mt-4 text-[var(--color-text-secondary)] leading-relaxed">
              FreStyle は、新卒・若手 IT エンジニアの研修を一気通貫で支える統合プラットフォームです。
              読むだけの座学で終わらせず、ブラウザ上のエディタで書いて即採点する演習まで含めることで、
              「わかる」だけでなく「できる」に届く研修を実現します。受講者は自分の伸びを、
              企業の研修担当はメンバーの学習状況を、それぞれ同じ場所で把握できます。
            </p>
          </div>
        </section>

        {/* 主な機能 */}
        <section className="mx-auto max-w-6xl px-6 py-16">
          <h2 className="text-2xl font-bold text-center">主な機能</h2>
          <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {FEATURES.map((feature) => (
              <div
                key={feature.title}
                className="rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] p-6"
              >
                <h3 className="text-lg font-semibold">{feature.title}</h3>
                <p className="mt-3 text-sm text-[var(--color-text-secondary)] leading-relaxed">
                  {feature.body}
                </p>
              </div>
            ))}
          </div>
        </section>

        {/* 導入の流れ */}
        <section className="bg-[var(--color-surface-1)] border-y border-[var(--color-surface-3)]">
          <div className="mx-auto max-w-6xl px-6 py-16">
            <h2 className="text-2xl font-bold text-center">導入の流れ</h2>
            <ol className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              {STEPS.map((step) => (
                <li
                  key={step.title}
                  className="rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface)] p-6"
                >
                  <h3 className="text-base font-semibold text-[var(--color-accent)]">{step.title}</h3>
                  <p className="mt-2 text-sm text-[var(--color-text-secondary)] leading-relaxed">
                    {step.body}
                  </p>
                </li>
              ))}
            </ol>
          </div>
        </section>

        {/* FAQ */}
        <section className="mx-auto max-w-4xl px-6 py-16">
          <h2 className="text-2xl font-bold text-center">よくある質問</h2>
          <dl className="mt-10 space-y-6">
            {FAQS.map((item) => (
              <div
                key={item.q}
                className="rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] p-6"
              >
                <dt className="font-semibold">{item.q}</dt>
                <dd className="mt-2 text-sm text-[var(--color-text-secondary)] leading-relaxed">
                  {item.a}
                </dd>
              </div>
            ))}
          </dl>
        </section>

        {/* 末尾 CTA */}
        <section className="bg-[var(--color-surface-1)] border-t border-[var(--color-surface-3)]">
          <div className="mx-auto max-w-4xl px-6 py-14 text-center">
            <h2 className="text-2xl font-bold">研修の実施と管理を、FreStyle で。</h2>
            <p className="mt-4 text-[var(--color-text-secondary)]">
              まずは企業の利用申請から。受講者の方は招待リンクからログインできます。
            </p>
            <div className="mt-8 flex flex-col sm:flex-row gap-3 justify-center">
              <Link
                to="/company-application"
                className="inline-flex items-center justify-center rounded-lg bg-[var(--color-accent)] px-6 py-3 font-semibold text-white shadow-sm hover:opacity-90 transition"
              >
                利用申請する
              </Link>
              <Link
                to="/login"
                className="inline-flex items-center justify-center rounded-lg border border-[var(--color-border-hover)] px-6 py-3 font-semibold text-[var(--color-text-primary)] hover:bg-[var(--color-surface-2)] transition"
              >
                ログイン
              </Link>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t border-[var(--color-surface-3)]">
        <div className="mx-auto max-w-6xl px-6 py-8 text-sm text-[var(--color-text-muted)] flex flex-col sm:flex-row items-center justify-between gap-3">
          <p>FreStyle — 新卒ITエンジニア向け研修プラットフォーム</p>
          <nav className="flex gap-5">
            <Link to="/login" className="hover:text-[var(--color-text-primary)]">
              ログイン
            </Link>
            <Link to="/company-application" className="hover:text-[var(--color-text-primary)]">
              利用申請
            </Link>
          </nav>
        </div>
      </footer>
    </div>
  );
}
