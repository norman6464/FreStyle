import { useEffect } from 'react';
import { Link, Navigate } from 'react-router-dom';
import {
  ArrowRightIcon,
  BookOpenIcon,
  ChartBarIcon,
  ChatBubbleLeftRightIcon,
  CheckCircleIcon,
  CheckIcon,
  CodeBracketIcon,
  MapIcon,
  PlusIcon,
  UserGroupIcon,
} from '@heroicons/react/24/outline';
import PublicHeader from '@/shared/ui/PublicHeader';
import { useAppSelector, useAppDispatch } from '@/shared/lib/store';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';
import { AuthRepository, setAuthData } from '@/entities/user';
import { clearAuthHintIfUnauthenticated } from '@/shared/lib/authHint';

const SITE_URL = 'https://frestyle.jp/';

const LANGUAGES = ['PHP', 'Java', 'JavaScript', 'TypeScript', 'Go', 'Ruby', 'C', 'C++', 'SQL'];

// 提供規模の実数（教材リポの seed 実績に合わせて更新する）。
const STATS: { value: string; unit: string; label: string }[] = [
  { value: '19', unit: 'コース', label: '実務直結のカリキュラム' },
  { value: '225', unit: '章', label: '章立てで途中から再開' },
  { value: '9', unit: '言語', label: 'ブラウザで書いて即採点' },
  { value: '4', unit: 'ステップ', label: '申請から研修開始まで' },
];

const FEATURES: { icon: typeof BookOpenIcon; title: string; body: string }[] = [
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

// ヒーローとフッターで共有する「方眼紙」モチーフ(エンジニアの学習ノート)。
const GRID_PATTERN_LIGHT =
  'bg-[linear-gradient(to_right,rgba(15,23,42,0.045)_1px,transparent_1px),linear-gradient(to_bottom,rgba(15,23,42,0.045)_1px,transparent_1px)] bg-[size:36px_36px]';
const GRID_PATTERN_DARK =
  'bg-[linear-gradient(to_right,rgba(255,255,255,0.05)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.05)_1px,transparent_1px)] bg-[size:36px_36px]';

/**
 * LandingPage は未ログイン/検索ボット向けの公開トップ。SEO のインデックス対象。
 * API は叩かず純静的に描画し、ビルド時プリレンダー(Playwright)で中身入り HTML を配信する。
 * ログイン済みで来た場合はダッシュボードへ送る（クライアント遷移時のみ。初回ロードは
 * 認証状態未確定なので LP を表示する = プリレンダーも LP になる）。
 *
 * 配色は Tailwind パレット（brand + 青バイアスの slate、正解表示のみ既存画面と同じ emerald）。
 * テーマ CSS 変数のうち accent 系は未定義で、参照すると透明背景になる事故が起きた（FRESTYLE-223）。
 * ヒーローの「タイプ → 採点 → 合格」演出は CSS アニメーションのみ（JS タイマー不使用）で、
 * prefers-reduced-motion では motion-reduce:* で静止画にフォールバックする。
 */
export default function LandingPage() {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const dispatch = useAppDispatch();

  // ログイン済みユーザーが直接 / を開いたときダッシュボードへ送るため、マウント時に
  // 一度だけ認証状態を確認する(Cookie は HttpOnly のため API に聞くしかない)。
  // 確認中・未ログイン(401)は LP をそのまま表示する(スピナーで隠すと SEO と初訪 UX を損なう)。
  // プリレンダー時はローカルサーバが /auth/me に 401 を返すので必ず LP が撮れる。
  useEffect(() => {
    if (isAuthenticated) return;
    let cancelled = false;
    AuthRepository.probeCurrentUser()
      .then((me) => {
        if (cancelled) return;
        dispatch(
          setAuthData({
            isAdmin: !!me.isAdmin,
            role: me.role ?? null,
            aiChatEnabledForTrainees: me.aiChatEnabledForTrainees ?? true,
          }),
        );
      })
      .catch((err) => {
        if (cancelled) return;
        // 未ログイン: LP のまま。認証切れが確定した(401/403)ときだけ目印を消す
        // (通信断で消すと、セッションが生きていても次回の振り分けが効かなくなる)。
        clearAuthHintIfUnauthenticated(err);
      });
    return () => {
      cancelled = true;
    };
    // マウント時に一度だけ確認する(isAuthenticated が true になったら <Navigate> が発火する)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
    <div className="h-full overflow-y-auto bg-white text-slate-900">
      <div className="sticky top-0 z-30">
        <PublicHeader />
      </div>

      <main>
        {/* ヒーロー: 「書いて即採点」がその場で動く */}
        <section className="relative overflow-hidden bg-[radial-gradient(52rem_30rem_at_82%_-12%,rgba(37,99,235,0.10),transparent_60%)]">
          <div
            aria-hidden="true"
            className={`pointer-events-none absolute inset-0 ${GRID_PATTERN_LIGHT} [mask-image:linear-gradient(to_bottom,black,transparent_78%)]`}
          />
          <div className="relative mx-auto max-w-6xl px-6 pt-20 pb-24 sm:pt-24">
            <div className="grid items-center gap-14 lg:grid-cols-2">
              <div className="text-center lg:text-left">
                <p className="inline-flex items-center gap-2 rounded-full border border-brand-200 bg-brand-50 px-3.5 py-1.5 text-xs font-bold text-brand-800">
                  <span aria-hidden="true" className="h-1.5 w-1.5 rounded-full bg-brand-600" />
                  企業向け・新卒エンジニア研修 SaaS
                </p>
                <h1 className="mt-6 text-[2.1rem] font-extrabold leading-[1.28] tracking-tight sm:text-[2.7rem] lg:text-[3rem]">
                  「わかる」で終わらせない、
                  <br />
                  <span className="text-brand-600">新卒ITエンジニア向け研修プラットフォーム</span>
                </h1>
                <p className="mx-auto mt-6 max-w-xl text-base leading-relaxed text-slate-600 sm:text-lg lg:mx-0">
                  コース学習・多言語のコーディング演習・AI チャット・学習レポートを1つに。
                  手を動かして学び、研修の進捗をまとめて管理できます。
                </p>
                <div className="mt-9 flex flex-col justify-center gap-3 sm:flex-row lg:justify-start">
                  <Link
                    to="/company-application"
                    className="inline-flex items-center justify-center gap-2 rounded-[10px] bg-brand-600 px-6 py-3 font-bold text-white shadow-[0_1px_2px_rgba(29,78,216,0.3),0_8px_24px_-8px_rgba(37,99,235,0.5)] transition hover:-translate-y-px hover:bg-brand-700"
                  >
                    企業の方：導入・利用申請
                    <ArrowRightIcon className="h-4 w-4" aria-hidden="true" />
                  </Link>
                  <Link
                    to="/login"
                    className="inline-flex items-center justify-center rounded-[10px] border border-slate-300 bg-white px-6 py-3 font-bold text-slate-800 transition hover:border-brand-200 hover:bg-brand-50"
                  >
                    受講者の方：ログイン
                  </Link>
                </div>
                <div className="mt-11">
                  <p className="text-[11.5px] font-bold tracking-[0.12em] text-slate-400">演習対応言語</p>
                  <ul className="mt-2.5 flex flex-wrap justify-center gap-1.5 lg:justify-start" aria-label="演習対応言語">
                    {LANGUAGES.map((lang) => (
                      <li
                        key={lang}
                        className="rounded-md border border-slate-200 bg-white px-2.5 py-0.5 font-mono text-xs text-slate-600"
                      >
                        {lang}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>

              {/* エディタ・ヴィネット: タイプ → 採点 → 合格の一連をモーションで見せる(装飾・CSS のみ) */}
              <div aria-hidden="true" className="relative hidden lg:block">
                <div className="rounded-2xl border border-slate-200 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.06),0_24px_60px_-24px_rgba(23,36,92,0.28)]">
                  <div className="flex items-center gap-1.5 border-b border-slate-200 px-5 py-3">
                    <span className="h-2.5 w-2.5 rounded-full bg-slate-200" />
                    <span className="h-2.5 w-2.5 rounded-full bg-slate-200" />
                    <span className="h-2.5 w-2.5 rounded-full bg-slate-200" />
                    <span className="ml-2.5 font-mono text-xs text-slate-400">コーディング演習 — greet.js</span>
                    <span className="ml-auto rounded-md border border-brand-200 bg-brand-50 px-2.5 py-0.5 text-[11.5px] font-bold text-brand-700">
                      ▶ 実行して採点
                    </span>
                  </div>
                  <pre className="overflow-hidden px-5 pt-5 pb-2 font-mono text-[13.5px] leading-[1.9] text-slate-700">
                    <code>
                      <span className="block opacity-0 animate-lp-type [animation-delay:0.5s] motion-reduce:animate-none motion-reduce:opacity-100">
                        <span className="mr-3 select-none text-slate-300">1</span>
                        <span className="font-semibold text-brand-700">function</span> greet(name) {'{'}
                      </span>
                      <span className="block opacity-0 animate-lp-type [animation-delay:1.1s] motion-reduce:animate-none motion-reduce:opacity-100">
                        <span className="mr-3 select-none text-slate-300">2</span>
                        {'  '}
                        <span className="font-semibold text-brand-700">return</span>{' '}
                        <span className="text-amber-700">{"'こんにちは、'"}</span> + name +{' '}
                        <span className="text-amber-700">{"' さん'"}</span>;
                      </span>
                      <span className="block opacity-0 animate-lp-type [animation-delay:1.6s] motion-reduce:animate-none motion-reduce:opacity-100">
                        <span className="mr-3 select-none text-slate-300">3</span>
                        {'}'}
                        <span className="ml-0.5 inline-block h-[15px] w-[7px] -mb-0.5 bg-brand-600 animate-lp-blink motion-reduce:animate-none" />
                      </span>
                    </code>
                  </pre>
                  <div className="mx-5 mb-5 origin-bottom rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3.5 opacity-0 [transform:translateY(6px)_scale(0.985)] animate-lp-stamp motion-reduce:animate-none motion-reduce:opacity-100 motion-reduce:[transform:none]">
                    <p className="flex items-center gap-2 text-sm font-bold text-emerald-700">
                      <CheckCircleIcon className="h-5 w-5" aria-hidden="true" />
                      正解 — 実行結果が期待出力と一致しました
                    </p>
                    <p className="mt-1 pl-7 font-mono text-xs text-emerald-800">こんにちは、FreStyle さん</p>
                  </div>
                </div>
                {/* 浮遊チップ: AI 質問と進捗の体験を添える */}
                <div className="absolute -right-3 -top-4 flex items-center gap-2.5 rounded-xl border border-slate-200 bg-white px-3.5 py-2.5 text-[12.5px] font-bold shadow-[0_12px_32px_-12px_rgba(23,36,92,0.25)] animate-lp-drift motion-reduce:animate-none">
                  <span className="grid h-7 w-7 place-items-center rounded-lg bg-brand-50 text-brand-700">
                    <ChatBubbleLeftRightIcon className="h-4 w-4" aria-hidden="true" />
                  </span>
                  詰まったら AI にその場で質問
                </div>
                <div className="absolute -bottom-5 -left-4 flex items-center gap-2.5 rounded-xl border border-slate-200 bg-white px-3.5 py-3 shadow-[0_12px_32px_-12px_rgba(23,36,92,0.25)] animate-lp-drift [animation-delay:1.2s] motion-reduce:animate-none">
                  <span className="block h-1.5 w-24 overflow-hidden rounded-full bg-slate-200">
                    <span className="block h-full w-[68%] rounded-full bg-brand-600" />
                  </span>
                  <span className="font-mono text-[11.5px] font-bold text-slate-400">Docker 入門 — 68%</span>
                </div>
              </div>
            </div>
          </div>

          {/* 統計バンド(提供規模の実数) */}
          <div className="relative border-y border-slate-200 bg-slate-50">
            <dl className="mx-auto grid max-w-6xl grid-cols-2 px-6 sm:grid-cols-4">
              {STATS.map((stat, index) => (
                <div
                  key={stat.label}
                  className={`px-4 py-7 text-center ${index > 0 ? 'sm:border-l sm:border-slate-200' : ''} ${
                    index % 2 === 1 ? 'border-l border-slate-200 sm:border-l' : ''
                  } ${index >= 2 ? 'border-t border-slate-200 sm:border-t-0' : ''}`}
                >
                  <dt className="order-2 mt-1 text-xs font-semibold text-slate-400">{stat.label}</dt>
                  <dd className="order-1 font-mono text-3xl font-extrabold tracking-tight [font-variant-numeric:tabular-nums]">
                    {stat.value}
                    <span className="ml-0.5 text-[15px] font-bold text-slate-600">{stat.unit}</span>
                  </dd>
                </div>
              ))}
            </dl>
          </div>
        </section>

        {/* FreStyle とは */}
        <section className="mx-auto grid max-w-6xl gap-10 px-6 py-24 lg:grid-cols-2 lg:gap-16">
          <div>
            <p className="flex items-center gap-2 text-xs font-bold tracking-[0.14em] text-brand-700">
              <span aria-hidden="true" className="h-0.5 w-5 bg-brand-600" />
              ABOUT
            </p>
            <h2 className="mt-3.5 text-[1.9rem] font-extrabold leading-snug tracking-tight">FreStyle とは</h2>
            <p className="mt-4 leading-relaxed text-slate-600">
              FreStyle は、新卒・若手 IT エンジニアの研修を一気通貫で支える統合プラットフォームです。
              読むだけの座学で終わらせず、ブラウザ上のエディタで書いて即採点する演習まで含めることで、
              「わかる」だけでなく「できる」に届く研修を実現します。受講者は自分の伸びを、
              企業の研修担当はメンバーの学習状況を、それぞれ同じ場所で把握できます。
            </p>
          </div>
          <ul className="space-y-3.5 self-center">
            {STRENGTHS.map((strength) => (
              <li
                key={strength}
                className="flex items-start gap-3.5 rounded-2xl border border-slate-200 bg-white px-5 py-5 font-semibold"
              >
                <span className="mt-0.5 grid h-7 w-7 flex-none place-items-center rounded-lg bg-emerald-50 text-emerald-600">
                  <CheckIcon className="h-4 w-4 [stroke-width:3]" aria-hidden="true" />
                </span>
                <span className="text-[15px] text-slate-800">{strength}</span>
              </li>
            ))}
          </ul>
        </section>

        {/* 主な機能(ベント配置: 学ぶ体験の 2 枚を主役に) */}
        <section className="border-y border-slate-200 bg-slate-50">
          <div className="mx-auto max-w-6xl px-6 py-24">
            <p className="flex items-center justify-center gap-2 text-xs font-bold tracking-[0.14em] text-brand-700">
              <span aria-hidden="true" className="h-0.5 w-5 bg-brand-600" />
              FEATURES
            </p>
            <h2 className="mt-3.5 text-center text-[1.9rem] font-extrabold tracking-tight sm:text-3xl">主な機能</h2>
            <p className="mt-3 text-center text-slate-600">研修の実施から進捗管理まで、必要なものを1つに。</p>

            <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-12">
              {/* 主役カード: コース学習(進捗のミニビジュアル付き) */}
              <div className="flex flex-col rounded-2xl border border-slate-200 bg-white p-7 transition hover:-translate-y-0.5 hover:border-brand-200 hover:shadow-[0_16px_40px_-20px_rgba(23,36,92,0.22)] lg:col-span-6">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-brand-50">
                  <BookOpenIcon className="h-6 w-6 text-brand-700" aria-hidden="true" />
                </span>
                <h3 className="mt-4 text-lg font-bold">コース学習</h3>
                <p className="mt-2 text-sm leading-relaxed text-slate-600">
                  Git・Docker・Linux・Go・PostgreSQL からクリーンアーキテクチャまで、実務直結のコースを章立てで学べる。進捗はコースごとに記録され、続きから再開できる。
                </p>
                <div aria-hidden="true" className="mt-auto grid gap-2.5 pt-6">
                  {[
                    { name: 'Git 入門', pct: 100 },
                    { name: 'Docker 入門', pct: 68 },
                    { name: 'Go 基礎', pct: 24 },
                  ].map((row) => (
                    <div
                      key={row.name}
                      className="grid grid-cols-[7.5em_1fr_3em] items-center gap-2.5 text-[11.5px] font-semibold text-slate-400"
                    >
                      <span>{row.name}</span>
                      <span className="block h-[7px] overflow-hidden rounded bg-slate-100">
                        <span className="block h-full rounded bg-brand-600" style={{ width: `${row.pct}%` }} />
                      </span>
                      <span className="text-right font-mono [font-variant-numeric:tabular-nums]">{row.pct}%</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* 主役カード: コーディング演習(ターミナルのミニビジュアル付き) */}
              <div className="flex flex-col rounded-2xl border border-slate-200 bg-white p-7 transition hover:-translate-y-0.5 hover:border-brand-200 hover:shadow-[0_16px_40px_-20px_rgba(23,36,92,0.22)] lg:col-span-6">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-brand-50">
                  <CodeBracketIcon className="h-6 w-6 text-brand-700" aria-hidden="true" />
                </span>
                <h3 className="mt-4 text-lg font-bold">コーディング演習</h3>
                <p className="mt-2 text-sm leading-relaxed text-slate-600">
                  ブラウザ上のエディタで書いて即採点。PHP・Java・JavaScript・TypeScript・Go・Ruby・C・C++・SQL など多言語に対応し、手を動かして身につける。
                </p>
                <div aria-hidden="true" className="mt-auto pt-6">
                  <div className="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 font-mono text-xs leading-[1.9] text-slate-700">
                    $ 提出しました…
                    <br />
                    <span className="font-bold text-emerald-600">✓ 正解</span> — 3 / 3 のテストに合格(0.12s)
                  </div>
                </div>
              </div>

              {FEATURES.map((feature) => (
                <div
                  key={feature.title}
                  className="rounded-2xl border border-slate-200 bg-white p-6 transition hover:-translate-y-0.5 hover:border-brand-200 hover:shadow-[0_16px_40px_-20px_rgba(23,36,92,0.22)] lg:col-span-3"
                >
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-brand-50">
                    <feature.icon className="h-6 w-6 text-brand-700" aria-hidden="true" />
                  </span>
                  <h3 className="mt-4 text-base font-bold">{feature.title}</h3>
                  <p className="mt-2 text-[13.5px] leading-relaxed text-slate-600">{feature.body}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* 導入の流れ(接続レール) */}
        <section className="mx-auto max-w-6xl px-6 py-24">
          <p className="flex items-center justify-center gap-2 text-xs font-bold tracking-[0.14em] text-brand-700">
            <span aria-hidden="true" className="h-0.5 w-5 bg-brand-600" />
            FLOW
          </p>
          <h2 className="mt-3.5 text-center text-[1.9rem] font-extrabold tracking-tight sm:text-3xl">導入の流れ</h2>
          <p className="mt-3 text-center text-slate-600">申請から研修開始まで、4 ステップで始められます。</p>
          <ol className="relative mt-14 grid gap-9 sm:grid-cols-2 lg:grid-cols-4">
            <span
              aria-hidden="true"
              className="absolute left-[12.5%] right-[12.5%] top-[19px] hidden h-0.5 bg-brand-100 lg:block"
            />
            {STEPS.map((step, index) => (
              <li key={step.title} className="relative px-2 text-center">
                <span className="relative z-[1] mx-auto grid h-10 w-10 place-items-center rounded-full bg-brand-600 font-mono text-[15px] font-extrabold text-white ring-[6px] ring-white">
                  {index + 1}
                </span>
                <h3 className="mt-4 text-base font-bold">{step.title}</h3>
                <p className="mt-2 text-[13.5px] leading-relaxed text-slate-600">{step.body}</p>
              </li>
            ))}
          </ol>
        </section>

        {/* FAQ(罫線リスト型) */}
        <section className="border-y border-slate-200 bg-slate-50">
          <div className="mx-auto max-w-6xl px-6 py-24">
            <p className="flex items-center justify-center gap-2 text-xs font-bold tracking-[0.14em] text-brand-700">
              <span aria-hidden="true" className="h-0.5 w-5 bg-brand-600" />
              FAQ
            </p>
            <h2 className="mt-3.5 text-center text-[1.9rem] font-extrabold tracking-tight sm:text-3xl">よくある質問</h2>
            <div className="mx-auto mt-11 max-w-3xl border-t border-slate-200">
              {FAQS.map((faq) => (
                <details key={faq.q} className="group border-b border-slate-200">
                  <summary className="flex cursor-pointer list-none items-center justify-between gap-5 px-1.5 py-6 font-bold [&::-webkit-details-marker]:hidden">
                    {faq.q}
                    <span
                      aria-hidden="true"
                      className="grid h-7 w-7 flex-none place-items-center rounded-lg bg-white text-brand-700 transition-transform group-open:rotate-45 group-open:bg-brand-50"
                    >
                      <PlusIcon className="h-4 w-4 [stroke-width:2.5]" />
                    </span>
                  </summary>
                  <p className="px-1.5 pb-6 pr-12 text-[14.5px] leading-relaxed text-slate-600">{faq.a}</p>
                </details>
              ))}
            </div>
          </div>
        </section>

        {/* CTA バンド(濃紺グラデーション + 方眼モチーフ) */}
        <section className="relative overflow-hidden bg-gradient-to-br from-[#17245c] via-[#1d3a8f] to-brand-700">
          <div aria-hidden="true" className={`pointer-events-none absolute inset-0 ${GRID_PATTERN_DARK}`} />
          <div className="relative mx-auto max-w-4xl px-6 py-20 text-center">
            <h2 className="text-[1.9rem] font-extrabold tracking-tight text-white sm:text-3xl">
              研修の実施と管理を、FreStyle で。
            </h2>
            <p className="mt-4 text-brand-100">
              まずは企業の利用申請から。受講者の方は招待リンクからログインできます。
            </p>
            <div className="mt-9 flex flex-col justify-center gap-3 sm:flex-row">
              <Link
                to="/company-application"
                className="inline-flex items-center justify-center gap-2 rounded-[10px] bg-white px-6 py-3 font-bold text-brand-700 shadow-[0_10px_30px_-10px_rgba(0,0,0,0.4)] transition hover:-translate-y-px hover:bg-brand-50"
              >
                企業の方：導入・利用申請
                <ArrowRightIcon className="h-4 w-4" aria-hidden="true" />
              </Link>
              <Link
                to="/login"
                className="inline-flex items-center justify-center rounded-[10px] border border-white/35 px-6 py-3 font-bold text-white transition hover:border-white/60 hover:bg-white/10"
              >
                受講者の方：ログイン
              </Link>
            </div>
          </div>
        </section>
      </main>

      {/* フッター */}
      <footer className="bg-slate-900">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-6 px-6 py-10 sm:flex-row">
          <div className="flex items-center gap-2">
            <img src="/favicon.svg" alt="" aria-hidden="true" className="h-6 w-6" />
            <span className="text-sm font-semibold text-white">
              FreStyle — 新卒ITエンジニア向け研修プラットフォーム
            </span>
          </div>
          <nav className="flex items-center gap-6 text-sm text-slate-400" aria-label="フッター">
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
