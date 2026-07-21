/**
 * FreStyle アプリで使う専門用語の定義集。
 * 画面内の GlossaryTerm / HelpTooltip で参照することで、
 * 文言の一貫性と説明の質を担保する。
 */
export const GLOSSARY = {
  fiveAxisScore: {
    term: '5軸評価',
    definition:
      '論理的構成力・配慮表現・要約力・提案力・質問/傾聴力の 5 つの観点で、会話内容をスコアリングする仕組みです。',
  },
  logicalStructure: {
    term: '論理的構成力',
    definition: '報連相を結論→理由→詳細の順で構造化できているか、という観点です。',
  },
  considerateExpression: {
    term: '配慮表現',
    definition: '敬語の正確さ、相手への配慮が込められた言い回しができているかを評価します。',
  },
  summarization: {
    term: '要約力',
    definition: '技術的な内容を、相手が理解しやすい平易な言葉でまとめられているかを評価します。',
  },
  proposalSkill: {
    term: '提案力',
    definition:
      'エスカレーション判断や代替案の提示など、状況に応じた建設的な提案ができているかを評価します。',
  },
  listeningSkill: {
    term: '質問・傾聴力',
    definition: '要件確認の網羅性と、相手の意図を汲み取る傾聴姿勢の両方を評価します。',
  },
  practiceMode: {
    term: '練習モード',
    definition:
      'AI がビジネスシーンのロールプレイ相手を演じ、12 種類のシナリオから選んで実践練習できる機能です。',
  },
  scenario: {
    term: 'シナリオ',
    definition:
      '障害報告、要件変更説明、見積もり交渉など、実務で遭遇する具体的なビジネス場面のことです。',
  },
  scoreCard: {
    term: 'スコアカード',
    definition: '会話終了後に AI が自動生成する、5 軸評価の結果と改善アドバイスのレポートです。',
  },
} as const;

export type GlossaryKey = keyof typeof GLOSSARY;
