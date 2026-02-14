import Card from './Card';

const TIPS = [
  { category: 'メール', text: '件名は「結論＋具体情報」で書くと開封率が上がります。例：「【承認依頼】○○プロジェクト予算案」' },
  { category: '会議', text: '発言の冒頭に「結論→理由→補足」の順で話すと、短い時間で的確に伝わります。' },
  { category: '報連相', text: '報告は「事実→影響→対応案」の3点セットで伝えると、上司の判断を助けられます。' },
  { category: 'チャット', text: 'テキストチャットでは感情が伝わりにくいため、ポジティブな表現を1つ添えると誤解を防げます。' },
  { category: 'メール', text: '長文メールは箇条書きに置き換えるだけで、読み手の理解速度が大幅に上がります。' },
  { category: '会議', text: '質問するときは「○○について確認したいのですが」と前置きすると、相手が答えやすくなります。' },
  { category: '報連相', text: '悪い報告こそ早めに。「まだ解決できていませんが、現状をお伝えします」が信頼を築きます。' },
  { category: 'チャット', text: '依頼メッセージには期限と優先度を明記すると、相手がスケジュールを立てやすくなります。' },
  { category: 'メール', text: '返信の冒頭で相手のメールの要点を一文で要約すると「ちゃんと読んでいる」と伝わります。' },
  { category: '会議', text: '議論が長引いたら「一度整理させてください」と切り出すと、場の流れをコントロールできます。' },
  { category: '報連相', text: '相談は「自分はAだと思うが、Bも検討したい」のように自分の意見を先に出すと建設的になります。' },
  { category: 'チャット', text: '「了解です」だけでなく「了解です、○○の件ですね」と一言添えると、認識ズレを防げます。' },
  { category: 'メール', text: 'CCに入れる人は「この情報を知っておくべき人」に絞ると、組織全体の生産性が上がります。' },
  { category: '会議', text: '会議の最後に「次のアクション」を全員で確認するだけで、タスクの抜け漏れが激減します。' },
];

function getTipOfTheDay(): (typeof TIPS)[number] {
  const today = new Date();
  const dayIndex = Math.floor(today.getTime() / (1000 * 60 * 60 * 24));
  return TIPS[dayIndex % TIPS.length];
}

export default function CommunicationTipCard() {
  const tip = getTipOfTheDay();

  return (
    <Card>
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-xs font-semibold text-[var(--color-text-primary)]">今日のTips</h3>
        <span
          data-testid="tip-category"
          className="text-[10px] font-medium text-amber-400 bg-amber-900/30 px-2 py-0.5 rounded"
        >
          {tip.category}
        </span>
      </div>
      <p data-testid="tip-text" className="text-sm text-[var(--color-text-tertiary)] leading-relaxed">
        {tip.text}
      </p>
    </Card>
  );
}
