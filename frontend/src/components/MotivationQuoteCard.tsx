import Card from './Card';

interface Quote {
  text: string;
  author: string;
}

const QUOTES: Quote[] = [
  { text: 'コミュニケーションで最も大切なのは、相手が言わなかったことを聞くことだ。', author: 'ピーター・ドラッカー' },
  { text: '話し上手は聞き上手。まず相手の話を聴くことから始めよう。', author: 'デール・カーネギー' },
  { text: '伝えることと伝わることは違う。伝わるまでが、コミュニケーション。', author: '松下幸之助' },
  { text: '言葉は、考えを伝える道具ではなく、関係を築く道具である。', author: 'ジョン・パウエル' },
  { text: '成功するチームの共通点は、心理的安全性が高いこと。', author: 'エイミー・エドモンドソン' },
  { text: '準備を怠ることは、失敗の準備をすることだ。', author: 'ベンジャミン・フランクリン' },
  { text: '質問する勇気が、成長への第一歩になる。', author: '稲盛和夫' },
  { text: 'フィードバックは贈り物。受け取る勇気が、あなたを強くする。', author: 'ケン・ブランチャード' },
  { text: '報連相は、チームの信頼を築く最も確実な方法である。', author: 'ビジネス格言' },
  { text: '今日の小さな一歩が、明日の大きな成長につながる。', author: '老子' },
];

function getDailyIndex(): number {
  const now = new Date();
  const dayOfYear = Math.floor(
    (now.getTime() - new Date(now.getFullYear(), 0, 0).getTime()) / 86400000
  );
  return dayOfYear % QUOTES.length;
}

export default function MotivationQuoteCard() {
  const quote = QUOTES[getDailyIndex()];

  return (
    <Card>
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-2">今日の一言</p>
      <p data-testid="quote-text" className="text-xs text-[var(--color-text-tertiary)] leading-relaxed italic">
        「{quote.text}」
      </p>
      <p data-testid="quote-author" className="text-[10px] text-[var(--color-text-faint)] mt-1.5 text-right">
        — {quote.author}
      </p>
    </Card>
  );
}
