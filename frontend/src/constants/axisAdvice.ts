/**
 * スキル軸別の改善アドバイス定数
 *
 * PracticeResultSummary と ScoreImprovementAdvice で共有
 */

export const IMPROVEMENT_ADVICE: Record<string, string> = {
  '論理的構成力': '結論→理由→具体例の順で話す練習をしましょう。報連相の構造化を意識してみてください。',
  '配慮表現': 'クッション言葉や敬語のバリエーションを増やしましょう。相手の立場に立った表現を心がけてください。',
  '要約力': '話す前に要点を3つに絞る練習をしましょう。技術的な内容を非エンジニアにも分かるように説明する訓練が効果的です。',
  '提案力': '課題だけでなく解決策をセットで伝える習慣をつけましょう。代替案を複数用意するとさらに効果的です。',
  '質問・傾聴力': '5W1Hを意識した質問を心がけましょう。相手の発言を要約して確認する「オウム返し」も有効です。',
};

export const SCORE_THRESHOLD = 7;

export function getAdviceForAxis(axis: string): string {
  return IMPROVEMENT_ADVICE[axis] || `${axis}を伸ばすための練習を続けましょう。`;
}
