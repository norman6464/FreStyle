interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

interface ScoreImprovementAdviceProps {
  scores: AxisScore[];
}

const AXIS_ADVICE: Record<string, string> = {
  '論理的構成力': '結論→理由→具体例の順で話す練習をしましょう。報連相の構造化を意識してみてください。',
  '配慮表現': 'クッション言葉や敬語のバリエーションを増やしましょう。相手の立場に立った表現を心がけてください。',
  '要約力': '話す前に要点を3つに絞る練習をしましょう。技術的な内容を非エンジニアにも分かるように説明する訓練が効果的です。',
  '提案力': '課題だけでなく解決策をセットで伝える習慣をつけましょう。代替案を複数用意するとさらに効果的です。',
  '質問・傾聴力': '5W1Hを意識した質問を心がけましょう。相手の発言を要約して確認する「オウム返し」も有効です。',
};

const THRESHOLD = 7;

export default function ScoreImprovementAdvice({ scores }: ScoreImprovementAdviceProps) {
  const weakAxes = scores.filter((s) => s.score < THRESHOLD);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[#D0D0D0] mb-3">改善アドバイス</p>

      {weakAxes.length === 0 ? (
        <p className="text-xs text-emerald-400 font-medium">
          素晴らしい成績です！この調子で続けましょう。
        </p>
      ) : (
        <div className="space-y-3">
          {weakAxes.map((axis) => (
            <div key={axis.axis} data-testid="advice-item" className="flex gap-3">
              <div className="flex-shrink-0 w-5 h-5 rounded-full bg-amber-100 flex items-center justify-center mt-0.5">
                <span className="text-amber-400 text-[10px] font-bold">!</span>
              </div>
              <div>
                <p className="text-xs font-medium text-[#F0F0F0]">
                  {axis.axis}（{axis.score.toFixed(1)}）
                </p>
                <p className="text-xs text-[#888888] mt-0.5">
                  {AXIS_ADVICE[axis.axis] || `${axis.axis}を伸ばすための練習を続けましょう。`}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
