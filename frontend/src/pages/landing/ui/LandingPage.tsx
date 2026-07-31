import { useEffect } from 'react';
import { Link, Navigate } from 'react-router-dom';
import {
  ArrowRightIcon,
  BookOpenIcon,
  ChartBarIcon,
  ChatBubbleLeftRightIcon,
  CheckCircleIcon,
  ChevronDownIcon,
  CodeBracketIcon,
  MapIcon,
  UserGroupIcon,
} from '@heroicons/react/24/outline';
import PublicHeader from '@/shared/ui/PublicHeader';
import { useAppSelector } from '@/shared/lib/store';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';

const SITE_URL = 'https://normanblog.com/';

const LANGUAGES = ['PHP', 'Java', 'JavaScript', 'TypeScript', 'Go', 'Ruby', 'C', 'C++', 'SQL'];

const FEATURES: { icon: typeof BookOpenIcon; title: string; body: string }[] = [
  {
    icon: BookOpenIcon,
    title: 'コース学習',
    body: 'Git・Docker・Linux・Go・PostgreSQL からクリーンアーキテクチャまで、実務直結のコースを章立てで学べる。進捗はコースごとに記録され、続きから再開できる。',
  },
  {
    icon: CodeBracketIcon,
    title: 'コーディング演習',
    body: 'ブラウザ上のエディタで書いて即採点。PHP・Java・JavaScript・TypeScript・Go・Ruby・C・C++・SQL など多言語に対応し、手を動かして身につける。',
  },
  {
    icon: ChatBubbleLeftRightIcon,
    title: 'AI チャットで質問',
    body: '詰まったところを AI にその場で相談。学習の流れを止めずに、疑問を解消しながら進められる。',
  },
  {
    icon: ChartBarIcon,
    title: '学習レポート',
    body: '取り組んだ量や進み具合をレポートで可視化。受講者は自分の伸びを、研修担当はチームの状況を把握できる。',
  },
  {
    icon: UserGroupIcon,
    title: 'メンバーの学習状況（管理者）',
    body: '企業の研修担当は、メンバーごとの進捗・つまずきを一覧で確認。招待もマジックリンクで簡単に行える。',
  },
  {
    icon: MapIcon,
    title: '迷わない導線',
    body: '「何を学ぶか」で迷わないよう、学習領域→コース→章の順に整理。初学者がひとりでも進められる設計。',
  },
];

const STRENGTHS: string[] = [
  '読むだけで終わらない — 書いて即採点する演習までが研修',
  '受講者の伸びと研修担当の把握を、同じ場所で',
  '初学者がひとりでも迷わず進める学習導線',
];

const STEPS: { title: string; body: string }[] = [
  { title: '利用申請', body: '企業の研修担当が利用申請を送信。折り返しご案内します。' },
  { title: 'メンバー招待', body: '受講者をマジックリンクで招待。受け取ったリンクから参加できます。' },
  { title: '学習開始', body: 'コース学習とコーディング演習で、手を動かしながら研修を進めます。' },
  { title: '進捗の確認', body: '学習レポートで、受講者・研修担当の双方が伸びを確認できます。' },
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
 *
 * 配色は tailwind.config.js 定義済みの brand / stone のみを使う（テーマ CSS 変数のうち
 * accent 系は未定義で、参照すると透明背景になる事故が起きた — FRESTYLE-223）。
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
    // body は overflow:hidden(AppShell の二重スクロールバー対策)のため、
    // LP 自身がスクロールコンテナを持つ(h-full + overflow-y-auto)。
    <div className="h-full overflow-y-auto bg-white text-stone-900">
      <div className="sticky top-0 z-30">
        <PublicHeader />
      </div>

      <main>
        {/* ヒーロー */}
        <section className="relative overflow-hidden bg-gradient-to-b from-brand-50 via-white to-white">
          <div className="mx-auto max-w-6xl px-6 pt-16 pb-20 sm:pt-24">
            <div className="grid items-center gap-12 lg:grid-cols-2">
              <div className="text-center lg:text-left">
                <p className="inline-flex items-center gap-1.5 rounded-full border border-brand-200 bg-brand-50 px-3 py-1 text-xs font-semibold text-brand-700">
                  企業向け・新卒エンジニア研修 SaaS
                </p>
                <h1 className="mt-5 text-3xl font-bold leading-tight tracking-tight sm:text-4xl lg:text-[2.6rem]">
                  「わかる」で終わらせない、
                  <br />
                  <span className="text-brand-600">新卒ITエンジニア向け研修プラットフォーム</span>
                </h1>
                <p className="mx-auto mt-6 max-w-xl text-base leading-relaxed text-stone-600 sm:text-lg lg:mx-0">
                  コース学習・多言語のコーディング演習・AI チャット・学習レポートを1つに。
                  手を動かして学び、研修の進捗をまとめて管理できます。
                </p>
                <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row lg:justify-start">
                  <Link
                    to="/company-application"
                    className="inline-flex items-center justify-center gap-2 rounded-lg bg-brand-600 px-6 py-3 font-semibold text-white shadow-sm transition hover:bg-brand-700"
                  >
                    企業の方：導入・利用申請
                    <ArrowRightIcon className="h-4 w-4" aria-hidden="true" />
                  </Link>
                  <Link
                    to="/login"
                    className="inline-flex items-center justify-center rounded-lg border border-stone-300 bg-white px-6 py-3 font-semibold text-stone-800 transition hover:bg-stone-50"
                  >
                    受講者の方：ログイン
                  </Link>
                </div>
                <ul className="mt-8 flex flex-wrap justify-center gap-1.5 lg:justify-start" aria-label="演習対応言語">
                  {LANGUAGES.map((lang) => (
                    <li
                      key={lang}
                      className="rounded-md border border-stone-200 bg-white px-2 py-0.5 font-mono text-xs text-stone-600"
                    >
                      {lang}
                    </li>
                  ))}
                </ul>
              </div>

              {/* プロダクトの体験を伝えるエディタ風モック(装飾・CSS のみ) */}
              <div aria-hidden="true" className="hidden lg:block">
                <div className="rounded-xl border border-stone-200 bg-white shadow-xl shadow-brand-100/60">
                  <div className="flex items-center gap-1.5 border-b border-stone-200 px-4 py-3">
                    <span className="h-2.5 w-2.5 rounded-full bg-stone-300" />
                    <span className="h-2.5 w-2.5 rounded-full bg-stone-300" />
                    <span className="h-2.5 w-2.5 rounded-full bg-stone-300" />
                    <span className="ml-3 text-xs font-medium text-stone-500">コーディング演習 — greet.js</span>
                  </div>
                  <pre className="overflow-hidden px-5 py-4 font-mono text-[13px] leading-6 text-stone-700">
                    <code>
                      {'function greet(name) {\n'}
                      {"  return 'こんにちは、' + name + ' さん';\n"}
                      {'}\n'}
                    </code>
                  </pre>
                  <div className="mx-5 mb-5 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3">
                    <p className="flex items-center gap-1.5 text-sm font-semibold text-emerald-700">
                      <CheckCircleIcon className="h-5 w-5" aria-hidden="true" />
                      正解 — 実行結果が期待出力と一致しました
                    </p>
                    <p className="mt-1 pl-6 font-mono text-xs text-emerald-800">こんにちは、FreStyle さん</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* FreStyle とは */}
        <section className="border-y border-stone-200 bg-stone-50">
          <div className="mx-auto grid max-w-6xl gap-10 px-6 py-16 lg:grid-cols-2 lg:gap-16">
            <div>
              <h2 className="text-2xl font-bold">FreStyle とは</h2>
              <p className="mt-4 leading-relaxed text-stone-600">
                FreStyle は、新卒・若手 IT エンジニアの研修を一気通貫で支える統合プラットフォームです。
                読むだけの座学で終わらせず、ブラウザ上のエディタで書いて即採点する演習まで含めることで、
                「わかる」だけでなく「できる」に届く研修を実現します。受講者は自分の伸びを、
                企業の研修担当はメンバーの学習状況を、それぞれ同じ場所で把握できます。
              </p>
            </div>
            <ul className="space-y-4 self-center">
              {STRENGTHS.map((strength) => (
                <li key={strength} className="flex items-start gap-3">
                  <CheckCircleIcon className="mt-0.5 h-6 w-6 shrink-0 text-brand-600" aria-hidden="true" />
                  <span className="font-medium text-stone-800">{strength}</span>
                </li>
              ))}
            </ul>
          </div>
        </section>

        {/* 主な機能 */}
        <section className="mx-auto max-w-6xl px-6 py-20">
          <h2 className="text-center text-2xl font-bold sm:text-3xl">主な機能</h2>
          <p className="mt-3 text-center text-stone-600">研修の実施から進捗管理まで、必要なものを1つに。</p>
          <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {FEATURES.map((feature) => (
              <div
                key={feature.title}
                className="rounded-xl border border-stone-200 bg-white p-6 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
              >
                <span className="inline-flex h-11 w-11 items-center justify-center rounded-lg bg-brand-50">
                  <feature.icon className="h-6 w-6 text-brand-600" aria-hidden="true" />
                </span>
                <h3 className="mt-4 text-lg font-semibold">{feature.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-stone-600">{feature.body}</p>
              </div>
            ))}
          </div>
        </section>

        {/* 導入の流れ */}
        <section className="border-y border-stone-200 bg-stone-50">
          <div className="mx-auto max-w-6xl px-6 py-20">
            <h2 className="text-center text-2xl font-bold sm:text-3xl">導入の流れ</h2>
            <p className="mt-3 text-center text-stone-600">申請から研修開始まで、4 ステップで始められます。</p>
            <ol className="mt-12 grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
              {STEPS.map((step, index) => (
                <li key={step.title} className="relative">
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-brand-600 text-base font-bold text-white">
                    {index + 1}
                  </span>
                  <h3 className="mt-4 text-base font-semibold">{step.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-stone-600">{step.body}</p>
                </li>
              ))}
            </ol>
          </div>
        </section>

        {/* FAQ */}
        <section className="mx-auto max-w-3xl px-6 py-20">
          <h2 className="text-center text-2xl font-bold sm:text-3xl">よくある質問</h2>
          <div className="mt-10 space-y-3">
            {FAQS.map((faq) => (
              <details
                key={faq.q}
                className="group rounded-xl border border-stone-200 bg-white open:shadow-sm"
              >
                <summary className="flex cursor-pointer list-none items-center justify-between gap-4 px-5 py-4 font-semibold [&::-webkit-details-marker]:hidden">
                  {faq.q}
                  <ChevronDownIcon
                    className="h-5 w-5 shrink-0 text-stone-400 transition-transform group-open:rotate-180"
                    aria-hidden="true"
                  />
                </summary>
                <p className="px-5 pb-5 text-sm leading-relaxed text-stone-600">{faq.a}</p>
              </details>
            ))}
          </div>
        </section>

        {/* CTA バンド */}
        <section className="bg-brand-600">
          <div className="mx-auto max-w-4xl px-6 py-16 text-center">
            <h2 className="text-2xl font-bold text-white sm:text-3xl">研修の実施と管理を、FreStyle で。</h2>
            <p className="mt-3 text-brand-100">
              まずは企業の利用申請から。受講者の方は招待リンクからログインできます。
            </p>
            <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
              <Link
                to="/company-application"
                className="inline-flex items-center justify-center gap-2 rounded-lg bg-white px-6 py-3 font-semibold text-brand-700 shadow-sm transition hover:bg-brand-50"
              >
                企業の方：導入・利用申請
                <ArrowRightIcon className="h-4 w-4" aria-hidden="true" />
              </Link>
              <Link
                to="/login"
                className="inline-flex items-center justify-center rounded-lg border border-brand-300 px-6 py-3 font-semibold text-white transition hover:bg-brand-700"
              >
                受講者の方：ログイン
              </Link>
            </div>
          </div>
        </section>
      </main>

      {/* フッター */}
      <footer className="bg-stone-900">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-6 px-6 py-10 sm:flex-row">
          <div className="flex items-center gap-2">
            <img src="/favicon.svg" alt="" aria-hidden="true" className="h-6 w-6" />
            <span className="text-sm font-semibold text-white">
              FreStyle — 新卒ITエンジニア向け研修プラットフォーム
            </span>
          </div>
          <nav className="flex items-center gap-6 text-sm text-stone-400" aria-label="フッター">
            <Link to="/login" className="transition hover:text-white">
              ログイン
            </Link>
            <Link to="/company-application" className="transition hover:text-white">
              利用申請
            </Link>
          </nav>
        </div>
      </footer>
    </div>
  );
}
