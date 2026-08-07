import { useEffect } from 'react';
import { Link, Navigate } from 'react-router-dom';
import {
  ArrowRightIcon,
  BellIcon,
  BookOpenIcon,
  ChartBarIcon,
  ChatBubbleLeftRightIcon,
  CheckCircleIcon,
  ClipboardDocumentListIcon,
  CodeBracketIcon,
  EnvelopeIcon,
  LockClosedIcon,
  MapIcon,
  PencilSquareIcon,
  ShieldCheckIcon,
  UserGroupIcon,
  UserIcon,
  UserPlusIcon,
} from '@heroicons/react/24/outline';
import PublicHeader from '@/shared/ui/PublicHeader';
import SectionHead from './SectionHead';
import CtaStrip from './CtaStrip';
import { useAppSelector, useAppDispatch } from '@/shared/lib/store';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';
import { AuthRepository, setAuthData } from '@/entities/user';
import { clearAuthHintIfUnauthenticated } from '@/shared/lib/authHint';

const SITE_URL = 'https://frestyle.jp/';

const LANGUAGES = ['PHP', 'Java', 'JavaScript', 'TypeScript', 'Go', 'Ruby', 'C', 'C++', 'SQL'];

// 機能カードのアイコン地色。5 色を循環させ、単色の羅列になるのを避ける。
const TINTS = {
  blue: 'bg-brand-50 text-brand-700',
  green: 'bg-emerald-50 text-emerald-600',
  violet: 'bg-violet-50 text-violet-600',
  amber: 'bg-amber-50 text-amber-600',
  teal: 'bg-teal-50 text-teal-600',
} as const;

const FEATURES: {
  icon: typeof BookOpenIcon;
  tint: keyof typeof TINTS;
  title: string;
  body: string;
}[] = [
  {
    icon: BookOpenIcon,
    tint: 'blue',
    title: 'コース学習',
    body: 'Git・Docker・Go・PostgreSQL などを章立てで学ぶ。続きから再開できる。',
  },
  {
    icon: CodeBracketIcon,
    tint: 'green',
    title: 'コーディング演習',
    body: 'ブラウザのエディタで書いて即採点。多言語に対応。',
  },
  {
    icon: ChatBubbleLeftRightIcon,
    tint: 'violet',
    title: 'AI チャット',
    body: '詰まったらその場で質問。学習の流れを止めない。',
  },
  {
    icon: ChartBarIcon,
    tint: 'amber',
    title: '学習レポート',
    body: '取り組んだ量と進み具合を可視化。伸びが見える。',
  },
  {
    icon: PencilSquareIcon,
    tint: 'teal',
    title: '学習ノート',
    body: '学びを自分の言葉で記録。画像も貼れる。',
  },
  {
    icon: UserGroupIcon,
    tint: 'blue',
    title: 'メンバーの学習状況',
    body: '研修担当は進捗・つまずきを一覧で確認できる。',
  },
  {
    icon: EnvelopeIcon,
    tint: 'green',
    title: 'メンバー招待',
    body: 'マジックリンクで招待。届いたリンクから参加するだけ。',
  },
  {
    icon: BellIcon,
    tint: 'violet',
    title: '通知',
    body: '申請や招待の動きを見逃さない。',
  },
  {
    icon: MapIcon,
    tint: 'amber',
    title: '迷わない導線',
    body: '学習領域→コース→章の順に整理。ひとりでも進められる。',
  },
];

const WALK_STEPS: { title: string; body: string }[] = [
  {
    title: '章を読んで理解する',
    body: '実務直結のコースを章立てで。読んだところまで自動で記録されます。',
  },
  {
    title: '演習を解いて、即採点',
    body: 'ブラウザのエディタで書いて提出。その場で採点され、「できた」が確かめられます。',
  },
  {
    title: '詰まったら AI に質問',
    body: '調べ物で学習が止まらない。文脈を踏まえた回答がその場で返ります。',
  },
  {
    title: 'レポートで振り返る',
    body: '取り組んだ量と伸びが自動で集計。研修担当も同じ画面でチームを見守れます。',
  },
];

const REPORT_ROWS: { name: string; pct: number }[] = [
  { name: 'Git 入門', pct: 100 },
  { name: 'Docker 入門', pct: 68 },
  { name: 'Go 基礎', pct: 24 },
  { name: 'SQL 演習', pct: 52 },
];

// Before/After 比較表。列 = 研修の観点、行 = 従来の研修 / FreStyle。
const COMPARISON_COLUMNS = ['教材と学び方', '質問対応', '進捗の把握'];
const COMPARISON_BEFORE: { heading: string; body: string }[] = [
  { heading: '読んで終わり', body: '座学とスライド中心。理解した「つもり」のまま配属日を迎えてしまう。' },
  {
    heading: '先輩の手が空くまで待つ',
    body: '質問のたびに先輩の作業が止まる。聞きづらくて詰まったまま進むことも。',
  },
  { heading: '日報と口頭で確認', body: '誰がどこで詰まっているかが見えず、フォローが後手に回る。' },
];
const COMPARISON_AFTER: { heading: string; body: string }[] = [
  {
    heading: '書いて、即採点',
    body: '章を読んだらすぐ演習。ブラウザで書いて採点され、「できる」まで届く。',
  },
  { heading: 'AI にその場で質問', body: '学習の流れを止めずに疑問を解消。先輩の時間を奪わない。' },
  {
    heading: 'レポートと一覧で見える',
    body: '受講者の伸びとつまずきを研修担当が同じ場所で把握。フォローが先手になる。',
  },
];

const ROLES: {
  icon: typeof UserIcon;
  tint: keyof typeof TINTS;
  label: string;
  title: string;
  items: string[];
}[] = [
  {
    icon: UserIcon,
    tint: 'blue',
    label: 'TRAINEE',
    title: '受講者',
    items: [
      'コースを読み、演習を解いて即採点',
      '詰まったら AI チャットに質問',
      'ノートに学びを記録、レポートで自分の伸びを確認',
    ],
  },
  {
    icon: UserGroupIcon,
    tint: 'green',
    label: 'COMPANY ADMIN',
    title: '企業の研修担当',
    items: [
      'メンバーをマジックリンクで招待',
      '進捗・つまずきを一覧で把握して先回りでフォロー',
      'AI 利用の可否などをメンバーごとに管理',
    ],
  },
  {
    icon: ShieldCheckIcon,
    tint: 'violet',
    label: 'OPERATOR',
    title: '運営',
    items: [
      '企業の利用申請を受け付けて開設',
      '会社・メンバーを横断で管理',
      '操作の監査ログで運用を透明に',
    ],
  },
];

const SECURITY: { icon: typeof ShieldCheckIcon; title: string; body: string }[] = [
  {
    icon: ShieldCheckIcon,
    title: '企業ごとのデータ分離',
    body: 'コース・演習・進捗は会社単位で分離して管理。',
  },
  {
    icon: UserPlusIcon,
    title: '招待制ログイン',
    body: '参加はマジックリンクの招待から。野良登録はできない。',
  },
  {
    icon: LockClosedIcon,
    title: 'ロール別の権限管理',
    body: '受講者・研修担当・運営で見える範囲と操作を分ける。',
  },
  {
    icon: ClipboardDocumentListIcon,
    title: '操作の監査ログ',
    body: '管理操作の記録を残し、運用を追跡できる。',
  },
];

// 濃紺ヒーロー・最終 CTA で共有する「方眼紙」モチーフ(エンジニアの学習ノート)。
const GRID_PATTERN_DARK =
  'bg-[linear-gradient(to_right,rgba(255,255,255,0.05)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.05)_1px,transparent_1px)] bg-[size:38px_38px]';

/**
 * LandingPage は未ログイン/検索ボット向けの公開トップ。SEO のインデックス対象。
 * API は叩かず純静的に描画し、ビルド時プリレンダー(Playwright)で中身入り HTML を配信する。
 * ログイン済みで来た場合はダッシュボードへ送る（クライアント遷移時のみ。初回ロードは
 * 認証状態未確定なので LP を表示する = プリレンダーも LP になる）。
 *
 * 構成は濃紺ヒーロー + 機能グリッド + 学習の進み方(ダーク) + Before/After 比較 +
 * ロール別 + セキュリティ + 反復 CTA。実数値・顧客ロゴ・事例は載せない方針。
 * 配色は Tailwind パレット（brand + slate、正解表示のみ既存画面と同じ emerald）。
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
        {/* ヒーロー(濃紺): コピー + 「書いて即採点」がその場で動く UI コラージュ */}
        <section className="relative overflow-hidden bg-[linear-gradient(165deg,#0e1c47_0%,#16295f_60%,#1d3a8f_100%)] text-white">
          <div
            aria-hidden="true"
            className={`pointer-events-none absolute inset-0 ${GRID_PATTERN_DARK} [mask-image:radial-gradient(60rem_40rem_at_30%_0%,black,transparent_75%)]`}
          />
          <div className="relative mx-auto max-w-6xl px-6 pb-24 pt-20 sm:pt-24">
            <div className="grid items-center gap-14 lg:grid-cols-2">
              <div className="text-center lg:text-left">
                <p className="inline-flex items-center gap-2 rounded-full border border-white/25 bg-white/10 px-3.5 py-1.5 text-xs font-bold text-[#dbe6ff]">
                  <span aria-hidden="true" className="h-1.5 w-1.5 rounded-full bg-[#8fb3ff]" />
                  企業向け・新卒エンジニア研修 SaaS
                </p>
                <h1 className="mt-6 text-[1.95rem] font-extrabold leading-[1.34] tracking-tight sm:text-[2.35rem] lg:text-[2.45rem]">
                  「わかる」で終わらせない、
                  <br />
                  <span className="text-[#8fb3ff]">
                    書いて、動かして、
                    <br />
                    「できる」に届く研修へ。
                  </span>
                </h1>
                <p className="mx-auto mt-6 max-w-xl text-base leading-relaxed text-[#c7d5f7] sm:text-lg lg:mx-0">
                  コース学習・多言語のコーディング演習・AI チャット・学習レポートを1つに。
                  手を動かして学び、研修の進捗をまとめて管理できます。
                </p>
                <div className="mt-9 flex flex-col justify-center gap-3 sm:flex-row lg:justify-start">
                  <Link
                    to="/company-application"
                    className="inline-flex items-center justify-center gap-2 rounded-full bg-white px-7 py-3 font-bold text-brand-700 shadow-[0_12px_32px_-12px_rgba(0,0,0,0.5)] transition hover:-translate-y-px hover:bg-brand-50"
                  >
                    企業の方：導入・利用申請
                    <ArrowRightIcon className="h-4 w-4" aria-hidden="true" />
                  </Link>
                  <Link
                    to="/login"
                    className="inline-flex items-center justify-center rounded-full border border-white/40 px-7 py-3 font-bold text-white transition hover:border-white/60 hover:bg-white/10"
                  >
                    受講者の方：ログイン
                  </Link>
                </div>
                <div className="mt-11">
                  <p className="text-[11.5px] font-bold tracking-[0.12em] text-[#8fa3d6]">演習対応言語</p>
                  <ul className="mt-2.5 flex flex-wrap justify-center gap-1.5 lg:justify-start" aria-label="演習対応言語">
                    {LANGUAGES.map((lang) => (
                      <li
                        key={lang}
                        className="rounded-md border border-white/20 bg-white/[.06] px-2.5 py-0.5 font-mono text-xs text-[#dbe6ff]"
                      >
                        {lang}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>

              {/* エディタ・ヴィネット: タイプ → 採点 → 合格の一連をモーションで見せる(装飾・CSS のみ) */}
              <div aria-hidden="true" className="relative hidden lg:block">
                <div className="rounded-2xl border border-white/20 bg-white text-slate-900 shadow-[0_30px_80px_-30px_rgba(0,0,0,0.6)]">
                  <div className="flex items-center gap-1.5 border-b border-slate-200 px-5 py-3">
                    <span className="h-2.5 w-2.5 rounded-full bg-slate-200" />
                    <span className="h-2.5 w-2.5 rounded-full bg-slate-200" />
                    <span className="h-2.5 w-2.5 rounded-full bg-slate-200" />
                    <span className="ml-2.5 font-mono text-xs text-slate-500">コーディング演習 — greet.js</span>
                  </div>
                  <pre className="overflow-hidden px-5 pb-2 pt-5 font-mono text-[13.5px] leading-[1.9] text-slate-700">
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
                        <span className="ml-0.5 -mb-0.5 inline-block h-[15px] w-[7px] bg-brand-600 animate-lp-blink motion-reduce:animate-none" />
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
                <div className="absolute -right-3 -top-4 flex items-center gap-2.5 rounded-xl border border-slate-200 bg-white px-3.5 py-2.5 text-[12.5px] font-bold text-slate-900 shadow-[0_16px_40px_-16px_rgba(0,0,0,0.5)] animate-lp-drift motion-reduce:animate-none">
                  <span className="grid h-7 w-7 place-items-center rounded-lg bg-brand-50 text-brand-700">
                    <ChatBubbleLeftRightIcon className="h-4 w-4" aria-hidden="true" />
                  </span>
                  詰まったら AI にその場で質問
                </div>
                <div className="absolute -bottom-5 -left-4 flex items-center gap-2.5 rounded-xl border border-slate-200 bg-white px-3.5 py-3 shadow-[0_16px_40px_-16px_rgba(0,0,0,0.5)] animate-lp-drift [animation-delay:1.2s] motion-reduce:animate-none">
                  <span className="block h-1.5 w-24 overflow-hidden rounded-full bg-slate-200">
                    <span className="block h-full w-[68%] rounded-full bg-brand-600" />
                  </span>
                  <span className="font-mono text-[11.5px] font-bold text-slate-500">Docker 入門 — 68%</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* 機能群グリッド: 小カード 9 枚で「必要な道具が揃っている」ことを見せる */}
        <section id="features" className="mx-auto max-w-6xl px-6 py-24">
          <SectionHead
            eyebrow="FEATURES"
            title="研修をつなぎ、学びの完了まで届ける機能群"
            lede="学ぶ・試す・振り返る・見守る。研修に必要な道具を、1つのプラットフォームに。"
          />
          <div className="mt-11 grid gap-3.5 sm:grid-cols-2 lg:grid-cols-3">
            {FEATURES.map((feature) => (
              <div
                key={feature.title}
                className="flex items-start gap-3.5 rounded-2xl border border-slate-200 bg-white px-[18px] py-[18px] transition hover:-translate-y-0.5 hover:border-brand-200 hover:shadow-[0_14px_36px_-20px_rgba(23,36,92,0.25)]"
              >
                <span
                  className={`grid h-[38px] w-[38px] flex-none place-items-center rounded-[10px] ${TINTS[feature.tint]}`}
                >
                  <feature.icon className="h-5 w-5" aria-hidden="true" />
                </span>
                <div>
                  <h3 className="text-[15px] font-bold">{feature.title}</h3>
                  <p className="mt-1 text-[12.8px] leading-[1.65] text-slate-600">{feature.body}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        <CtaStrip text="受講者の学習は 4 ステップで進む" />

        {/* 学習の進み方(ダーク): 4 ステップ + 学習レポートのモック */}
        <section id="flow" className="bg-[linear-gradient(180deg,#0e1c47_0%,#16295f_100%)]">
          <div className="mx-auto max-w-6xl px-6 py-24">
            <SectionHead
              dark
              eyebrow="HOW IT WORKS"
              title="学びっぱなしにさせない。進捗はすべて FreStyle が記録する"
              lede="受講者は学ぶことに集中するだけ。記録と可視化はプラットフォームの仕事です。"
            />
            <div className="mt-12 grid items-center gap-12 lg:grid-cols-2">
              <ol className="border-t border-white/[.14]">
                {WALK_STEPS.map((step, index) => (
                  <li key={step.title} className="flex gap-4 border-b border-white/[.14] px-1 py-5">
                    <span className="mt-0.5 grid h-[30px] w-[30px] flex-none place-items-center rounded-full border border-white/25 bg-white/10 font-mono text-[13px] font-extrabold text-[#9db8f5]">
                      {index + 1}
                    </span>
                    <div>
                      <h3 className="text-base font-bold text-white">{step.title}</h3>
                      <p className="mt-1 text-[13.5px] leading-relaxed text-[#c7d5f7]">{step.body}</p>
                    </div>
                  </li>
                ))}
              </ol>
              <div
                aria-hidden="true"
                className="overflow-hidden rounded-2xl border border-white/20 bg-white text-slate-900 shadow-[0_30px_80px_-30px_rgba(0,0,0,0.6)]"
              >
                <div className="flex items-center justify-between border-b border-slate-200 px-[18px] py-3.5">
                  <span className="text-[13px] font-bold">学習レポート</span>
                  <span className="rounded-md bg-brand-50 px-2.5 py-0.5 text-[11px] font-bold text-brand-700">
                    今週の進捗
                  </span>
                </div>
                <div className="grid gap-2.5 p-[18px]">
                  {REPORT_ROWS.map((row) => (
                    <div
                      key={row.name}
                      className="grid grid-cols-[7.5em_1fr_3.2em] items-center gap-2.5 text-xs font-semibold text-slate-500"
                    >
                      <span>{row.name}</span>
                      <span className="block h-[7px] overflow-hidden rounded bg-slate-50">
                        <span className="block h-full rounded bg-brand-600" style={{ width: `${row.pct}%` }} />
                      </span>
                      <span className="text-right font-mono [font-variant-numeric:tabular-nums]">{row.pct}%</span>
                    </div>
                  ))}
                </div>
                <div className="mx-[18px] mb-[18px] rounded-[10px] border border-emerald-200 bg-emerald-50 px-3.5 py-2.5 text-[12.5px] font-bold text-emerald-700">
                  ✓ 今日も学習を継続中 — 演習 3 問に合格
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Before / After 比較: 従来の研修との違いを観点別に見せる */}
        <section id="compare" className="mx-auto max-w-6xl px-6 py-24">
          <SectionHead
            eyebrow="BEFORE / AFTER"
            title="FreStyle で変わる新人研修"
            lede="「教える・答える・把握する」の手間を、プラットフォームに任せる。"
          />
          {/* 表は狭い画面で横スクロールするため、キーボードでも操作できるようフォーカス可能にする */}
          <div
            role="region"
            aria-label="従来の研修と FreStyle の比較"
            tabIndex={0}
            className="mt-11 overflow-x-auto rounded-[18px] border border-slate-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-600"
          >
            <table className="w-full min-w-[640px] border-collapse text-sm">
              <thead>
                <tr className="bg-slate-50 text-left text-[13px] text-slate-500">
                  <th scope="col" className="w-[9em] border-b border-slate-200 px-5 py-3.5" />
                  {COMPARISON_COLUMNS.map((col) => (
                    <th
                      key={col}
                      scope="col"
                      className="border-b border-l border-slate-200 px-5 py-3.5 font-bold"
                    >
                      {col}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                <tr className="align-top">
                  <th
                    scope="row"
                    className="border-b border-slate-200 bg-slate-50 px-5 py-5 text-left text-[13.5px] font-bold text-slate-500"
                  >
                    従来の研修
                  </th>
                  {COMPARISON_BEFORE.map((cell) => (
                    <td key={cell.heading} className="border-b border-l border-slate-200 px-5 py-5">
                      <span className="block text-[15px] font-extrabold text-slate-500">{cell.heading}</span>
                      <p className="mt-1 text-[13px] leading-relaxed text-slate-600">{cell.body}</p>
                    </td>
                  ))}
                </tr>
                <tr className="align-top">
                  <th
                    scope="row"
                    className="bg-slate-50 px-5 py-5 text-left text-[13.5px] font-bold text-slate-500"
                  >
                    FreStyle
                  </th>
                  {COMPARISON_AFTER.map((cell) => (
                    <td key={cell.heading} className="border-l border-slate-200 bg-brand-50 px-5 py-5">
                      <span className="block text-[15px] font-extrabold text-emerald-600">
                        ✓ {cell.heading}
                      </span>
                      <p className="mt-1 text-[13px] leading-relaxed text-slate-600">{cell.body}</p>
                    </td>
                  ))}
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <CtaStrip text="受講者は招待リンクから参加するだけ" />

        {/* ロール別の使い方: 立場ごとの体験を並べる */}
        <section className="mx-auto max-w-6xl px-6 pb-24 pt-10">
          <SectionHead eyebrow="FOR EACH ROLE" title="立場ごとに、見える景色を用意" />
          <div className="mt-11 grid gap-4 lg:grid-cols-3">
            {ROLES.map((role) => (
              <div key={role.title} className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
                <div className="flex items-center gap-3 border-b border-slate-200 bg-slate-50 px-5 py-4">
                  <span
                    className={`grid h-9 w-9 flex-none place-items-center rounded-[10px] ${TINTS[role.tint]}`}
                  >
                    <role.icon className="h-5 w-5" aria-hidden="true" />
                  </span>
                  <div>
                    <p className="text-[11px] font-bold tracking-[0.06em] text-slate-500">{role.label}</p>
                    <h3 className="text-[15.5px] font-extrabold">{role.title}</h3>
                  </div>
                </div>
                <ul className="grid gap-2.5 px-5 pb-5 pt-4">
                  {role.items.map((item) => (
                    <li key={item} className="relative pl-[22px] text-[13.5px] text-slate-600">
                      <span
                        aria-hidden="true"
                        className="absolute left-0 top-[7px] h-3 w-3 rounded border-[1.5px] border-brand-200 bg-brand-50"
                      />
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </section>

        {/* セキュリティと運用: 企業導入の不安に先回りで答える */}
        <section className="border-y border-slate-200 bg-slate-50">
          <div className="mx-auto max-w-6xl px-6 py-20">
            <SectionHead eyebrow="SECURITY" title="企業利用を前提にした設計" />
            <div className="mt-11 grid gap-3.5 sm:grid-cols-2 lg:grid-cols-4">
              {SECURITY.map((item) => (
                <div key={item.title} className="rounded-2xl border border-slate-200 bg-white p-5 text-center">
                  <span className="mx-auto grid h-11 w-11 place-items-center rounded-xl bg-brand-50 text-brand-700">
                    <item.icon className="h-[22px] w-[22px]" aria-hidden="true" />
                  </span>
                  <h3 className="mt-3 text-[14.5px] font-bold">{item.title}</h3>
                  <p className="mt-1.5 text-[12.5px] leading-relaxed text-slate-600">{item.body}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* 最終 CTA(濃紺 + 方眼モチーフ) */}
        <section className="relative overflow-hidden bg-[linear-gradient(160deg,#0e1c47_0%,#1d3a8f_70%,#1d4ed8_100%)]">
          <div aria-hidden="true" className={`pointer-events-none absolute inset-0 ${GRID_PATTERN_DARK}`} />
          <div className="relative mx-auto max-w-4xl px-6 py-20 text-center">
            <h2 className="text-[1.9rem] font-extrabold tracking-tight text-white sm:text-3xl">
              研修の実施と管理を、FreStyle で。
            </h2>
            <p className="mt-4 text-[#c7d5f7]">
              まずは企業の利用申請から。受講者の方は招待リンクからログインできます。
            </p>
            <div className="mt-9 flex flex-col justify-center gap-3 sm:flex-row">
              <Link
                to="/company-application"
                className="inline-flex items-center justify-center gap-2 rounded-full bg-white px-7 py-3 font-bold text-brand-700 shadow-[0_12px_32px_-12px_rgba(0,0,0,0.5)] transition hover:-translate-y-px hover:bg-brand-50"
              >
                企業の方：導入・利用申請
                <ArrowRightIcon className="h-4 w-4" aria-hidden="true" />
              </Link>
              <Link
                to="/login"
                className="inline-flex items-center justify-center rounded-full border border-white/40 px-7 py-3 font-bold text-white transition hover:border-white/60 hover:bg-white/10"
              >
                受講者の方：ログイン
              </Link>
            </div>
          </div>
        </section>
      </main>

      {/* フッター */}
      <footer className="bg-[#0a1330]">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-6 px-6 py-10 sm:flex-row">
          <div className="flex items-center gap-2">
            <img src="/favicon.svg" alt="" aria-hidden="true" className="h-6 w-6" />
            <span className="text-sm font-semibold text-white">
              FreStyle — 新卒ITエンジニア向け研修プラットフォーム
            </span>
          </div>
          <nav className="flex items-center gap-6 text-sm text-[#9aa5c4]" aria-label="フッター">
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
