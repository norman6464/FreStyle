import { Link } from 'react-router-dom';
import {
  AcademicCapIcon,
  ChatBubbleLeftRightIcon,
  ChartBarIcon,
  DocumentTextIcon,
  LifebuoyIcon,
  QuestionMarkCircleIcon,
  RocketLaunchIcon,
  SparklesIcon,
  BookmarkIcon,
} from '@heroicons/react/24/outline';
import { PageIntro, StepIndicator, GuidedHint, GlossaryTerm, ActionCard } from '../components/ui';
import { GLOSSARY } from '../constants/glossary';

/**
 * 新卒新入社員向けの「使い方ガイド」ページ。
 *
 * 章立て:
 *   1. FreStyle ってなに？
 *   2. 最初の1日にやること（StepIndicator + ActionCard）
 *   3. 練習モードの使い方（シナリオの選び方）
 *   4. 5軸評価の読み方
 *   5. AI チャットとは
 *   6. メモ・テンプレート機能
 *   7. 困ったとき (FAQ)
 *
 * - 既存の共通UIコンポーネント（PageIntro / StepIndicator / GlossaryTerm /
 *   GuidedHint / ActionCard）を最大限活用し、文章だけでなく操作可能な
 *   導線まで含めて「読んだあと即実践できる」ことを最優先にする。
 * - <h2>/<h3> によるアウトライン構造を意識し、aria-labelledby で
 *   セクションをスクリーンリーダーに伝える。
 */
export default function HelpPage() {
  return (
    <div className="p-6 max-w-3xl mx-auto space-y-10 pb-16">
      <PageIntro
        icon={<LifebuoyIcon className="h-6 w-6" />}
        title="使い方ガイド"
        description="新卒・新入社員の方が、初日から迷わず FreStyle を使い始められるようにまとめた入門ドキュメントです。"
      />

      <GuidedHint title="このページの読み方" storageKey="hint:help:howto-v1">
        まずは「最初の1日にやること」を順番に試してみてください。約 10〜15 分でアプリの主要機能を一通り体験できます。
      </GuidedHint>

      {/* 1. FreStyle ってなに？ */}
      <section aria-labelledby="help-what">
        <h2 id="help-what" className="mb-3 text-xl font-bold text-[var(--color-text-primary)]">
          1. FreStyle ってなに？
        </h2>
        <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
          FreStyle は、新卒 IT エンジニアのための <strong>ビジネスコミュニケーション練習アプリ</strong> です。
          <br />
          顧客折衝・上司への報連相・設計レビューでのやり取りなど、実務で遭遇しがちな 12 種類のシーンを AI 相手に
          ロールプレイし、AI のフィードバックでコミュニケーション力を伸ばせます。
        </p>
        <ul className="mt-3 list-disc pl-6 text-sm text-[var(--color-text-secondary)] space-y-1">
          <li>
            <GlossaryTerm
              term={GLOSSARY.scenario.term}
              definition={GLOSSARY.scenario.definition}
            />{' '}
            を選んで AI と会話する
          </li>
          <li>
            会話終了後に{' '}
            <GlossaryTerm
              term={GLOSSARY.scoreCard.term}
              definition={GLOSSARY.scoreCard.definition}
            />{' '}
            が自動生成され、改善ポイントが見える
          </li>
          <li>
            <GlossaryTerm
              term={GLOSSARY.fiveAxisScore.term}
              definition={GLOSSARY.fiveAxisScore.definition}
            />{' '}
            で自分の強み・弱みを可視化できる
          </li>
        </ul>
      </section>

      {/* 2. 最初の1日にやること */}
      <section aria-labelledby="help-firstday">
        <h2 id="help-firstday" className="mb-3 text-xl font-bold text-[var(--color-text-primary)]">
          2. 最初の1日にやること
        </h2>
        <p className="mb-4 text-sm text-[var(--color-text-secondary)] leading-relaxed">
          以下の 3 ステップを順番に試すと、アプリの主要機能を一通り体験できます。
        </p>
        <div className="mb-4">
          <StepIndicator
            steps={[
              { label: 'シナリオを選ぶ', description: '12 件から 1 つ' },
              { label: 'AI と会話する', description: '5〜10 分のロールプレイ' },
              { label: 'スコアで振り返る', description: '5 軸評価を確認' },
            ]}
            currentStep={0}
          />
        </div>
        <div className="grid gap-3 sm:grid-cols-1">
          <ActionCard
            to="/practice"
            title="練習モードを開く"
            description="まずはお勧めシナリオから 1 つ選んで会話を始めてみましょう。"
            icon={<AcademicCapIcon className="h-5 w-5" />}
            emphasis="primary"
            badge="Step 1"
          />
          <ActionCard
            to="/scores"
            title="スコア履歴を確認する"
            description="練習が終わったら、5軸評価の結果と推移を見てみましょう。"
            icon={<ChartBarIcon className="h-5 w-5" />}
            badge="Step 3"
          />
        </div>
      </section>

      {/* 3. 練習モードの使い方 */}
      <section aria-labelledby="help-practice">
        <h2 id="help-practice" className="mb-3 text-xl font-bold text-[var(--color-text-primary)]">
          3.{' '}
          <GlossaryTerm
            term={GLOSSARY.practiceMode.term}
            definition={GLOSSARY.practiceMode.definition}
          />{' '}
          の使い方
        </h2>
        <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
          12 種類のビジネスシーンから、いま伸ばしたい力に合うシナリオを選んで AI と会話します。
        </p>

        <h3 className="mt-4 text-base font-semibold text-[var(--color-text-primary)]">
          シナリオの選び方（迷ったら）
        </h3>
        <ul className="mt-2 list-disc pl-6 text-sm text-[var(--color-text-secondary)] space-y-1">
          <li>
            <strong>はじめての人</strong>: 「日報を上司に共有する」など報連相系から
          </li>
          <li>
            <strong>顧客対応に不安</strong>: 「障害報告」「要件変更の影響説明」を選ぶ
          </li>
          <li>
            <strong>シニアとの対話練習</strong>: 「設計レビューでの意見対立」「コードレビューの指摘対応」
          </li>
          <li>
            ホーム画面の「おすすめシナリオ」は、あなたの直近の弱み（最も低かった軸）から逆算して提案されます
          </li>
        </ul>

        <h3 className="mt-4 text-base font-semibold text-[var(--color-text-primary)]">
          会話のコツ
        </h3>
        <ul className="mt-2 list-disc pl-6 text-sm text-[var(--color-text-secondary)] space-y-1">
          <li>「結論 → 理由 → 詳細」の順で書くと論理的構成力スコアが伸びやすい</li>
          <li>分からないことは素直に質問すると質問・傾聴力が評価される</li>
          <li>専門用語は相手に応じて噛み砕く（要約力）</li>
        </ul>
      </section>

      {/* 4. 5軸評価の読み方 */}
      <section aria-labelledby="help-scoring">
        <h2 id="help-scoring" className="mb-3 text-xl font-bold text-[var(--color-text-primary)]">
          4.{' '}
          <GlossaryTerm
            term={GLOSSARY.fiveAxisScore.term}
            definition={GLOSSARY.fiveAxisScore.definition}
          />{' '}
          の読み方
        </h2>
        <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
          会話終了後に AI が以下の 5 つの観点で自動採点します。スコアは 0〜10 点で、各軸ごとに改善コメントが付きます。
        </p>
        <dl className="mt-3 grid gap-3 sm:grid-cols-2">
          <div className="rounded-lg border border-surface-3 bg-surface-1 p-3">
            <dt className="text-sm font-semibold text-primary-300">
              {GLOSSARY.logicalStructure.term}
            </dt>
            <dd className="mt-1 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              {GLOSSARY.logicalStructure.definition}
            </dd>
          </div>
          <div className="rounded-lg border border-surface-3 bg-surface-1 p-3">
            <dt className="text-sm font-semibold text-primary-300">
              {GLOSSARY.considerateExpression.term}
            </dt>
            <dd className="mt-1 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              {GLOSSARY.considerateExpression.definition}
            </dd>
          </div>
          <div className="rounded-lg border border-surface-3 bg-surface-1 p-3">
            <dt className="text-sm font-semibold text-primary-300">
              {GLOSSARY.summarization.term}
            </dt>
            <dd className="mt-1 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              {GLOSSARY.summarization.definition}
            </dd>
          </div>
          <div className="rounded-lg border border-surface-3 bg-surface-1 p-3">
            <dt className="text-sm font-semibold text-primary-300">
              {GLOSSARY.proposalSkill.term}
            </dt>
            <dd className="mt-1 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              {GLOSSARY.proposalSkill.definition}
            </dd>
          </div>
          <div className="rounded-lg border border-surface-3 bg-surface-1 p-3 sm:col-span-2">
            <dt className="text-sm font-semibold text-primary-300">
              {GLOSSARY.listeningSkill.term}
            </dt>
            <dd className="mt-1 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              {GLOSSARY.listeningSkill.definition}
            </dd>
          </div>
        </dl>
        <p className="mt-3 text-xs text-[var(--color-text-muted)] leading-relaxed">
          スコアは 1 回の会話だけで判断せず、5〜10 セッションを目安に推移を見るのがおすすめです。ホームの「成長トレンド」で確認できます。
        </p>
      </section>

      {/* 5. AI チャットとは */}
      <section aria-labelledby="help-aichat">
        <h2 id="help-aichat" className="mb-3 text-xl font-bold text-[var(--color-text-primary)]">
          5. AI チャットとは
        </h2>
        <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
          練習モード以外に、AI に対して自由に質問・相談できる「AI アシスタント」機能があります。
        </p>
        <ul className="mt-2 list-disc pl-6 text-sm text-[var(--color-text-secondary)] space-y-1">
          <li>「この報告メールどう書けば良い？」と書きかけの文面を相談する</li>
          <li>「明日の打ち合わせで何を聞けばいい？」と論点整理を依頼する</li>
          <li>過去のチャットは履歴として残るので、振り返りや見直しに使える</li>
        </ul>
        <div className="mt-3">
          <ActionCard
            to="/chat/ask-ai"
            title="AI アシスタントに相談する"
            description="自由形式で質問・壁打ちができます。"
            icon={<SparklesIcon className="h-5 w-5" />}
          />
        </div>
      </section>

      {/* 6. メモ・テンプレート */}
      <section aria-labelledby="help-notes">
        <h2 id="help-notes" className="mb-3 text-xl font-bold text-[var(--color-text-primary)]">
          6. メモ・テンプレート機能
        </h2>
        <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
          学んだフレーズや繰り返し使う文面は、メモ・テンプレート機能に保存できます。
        </p>
        <div className="mt-3 grid gap-3 sm:grid-cols-2">
          <ActionCard
            to="/notes"
            title="メモを書く"
            description="気付き・改善ポイントを Markdown で残せます。"
            icon={<DocumentTextIcon className="h-5 w-5" />}
          />
          <ActionCard
            to="/templates"
            title="テンプレートを使う"
            description="「障害報告」など定型のひな形を再利用。"
            icon={<BookmarkIcon className="h-5 w-5" />}
          />
        </div>
      </section>

      {/* 7. FAQ */}
      <section aria-labelledby="help-faq">
        <h2 id="help-faq" className="mb-3 text-xl font-bold text-[var(--color-text-primary)]">
          7. 困ったとき（FAQ）
        </h2>
        <dl className="space-y-3">
          <details className="rounded-lg border border-surface-3 bg-surface-1 p-3 open:bg-surface-2/40">
            <summary className="cursor-pointer text-sm font-semibold text-[var(--color-text-primary)]">
              <QuestionMarkCircleIcon className="mr-1 inline h-4 w-4 text-primary-300" aria-hidden="true" />
              スコアが低くて落ち込みます
            </summary>
            <p className="mt-2 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              はじめは誰でも 5 点前後からスタートします。3 セッション目以降から伸びてくる人が多いので、まずは 1 週間続けることを目標にしてみてください。
            </p>
          </details>
          <details className="rounded-lg border border-surface-3 bg-surface-1 p-3 open:bg-surface-2/40">
            <summary className="cursor-pointer text-sm font-semibold text-[var(--color-text-primary)]">
              <QuestionMarkCircleIcon className="mr-1 inline h-4 w-4 text-primary-300" aria-hidden="true" />
              どのシナリオから始めればいい？
            </summary>
            <p className="mt-2 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              ホーム画面の「おすすめシナリオ」をまず開いてみてください。あなたの最も低い軸に効くシナリオが提示されます。それでも迷う場合は「日報を共有する」など軽めのシナリオから。
            </p>
          </details>
          <details className="rounded-lg border border-surface-3 bg-surface-1 p-3 open:bg-surface-2/40">
            <summary className="cursor-pointer text-sm font-semibold text-[var(--color-text-primary)]">
              <QuestionMarkCircleIcon className="mr-1 inline h-4 w-4 text-primary-300" aria-hidden="true" />
              AI の返答がしっくり来ないとき
            </summary>
            <p className="mt-2 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              AI は完璧ではありません。あくまで「練習相手」として割り切り、納得いかない返答が来たら「もっと具体的に教えて」と聞き返してみましょう。それ自体が質問・傾聴力の練習になります。
            </p>
          </details>
          <details className="rounded-lg border border-surface-3 bg-surface-1 p-3 open:bg-surface-2/40">
            <summary className="cursor-pointer text-sm font-semibold text-[var(--color-text-primary)]">
              <QuestionMarkCircleIcon className="mr-1 inline h-4 w-4 text-primary-300" aria-hidden="true" />
              続かない / 気が乗らない
            </summary>
            <p className="mt-2 text-xs text-[var(--color-text-secondary)] leading-relaxed">
              1 日 1 回・5 分でも構いません。ホームの「日次目標」「ストリークカレンダー」が継続をサポートします。週 3 回を 4 週間続けると効果を実感しやすいです。
            </p>
          </details>
        </dl>
      </section>

      {/* CTA: ホームへ戻る */}
      <section aria-label="次のアクション">
        <ActionCard
          to="/"
          title="ホームに戻って練習を始める"
          description="読み終わったら、まずは 1 セッション体験してみましょう。"
          icon={<RocketLaunchIcon className="h-5 w-5" />}
          emphasis="primary"
        />
        <p className="mt-3 flex items-center gap-1 text-xs text-[var(--color-text-muted)]">
          <ChatBubbleLeftRightIcon className="h-3.5 w-3.5" aria-hidden="true" />
          このガイドに載っていない疑問は、AI アシスタントに直接質問するのが早いです。
        </p>
      </section>
    </div>
  );
}
