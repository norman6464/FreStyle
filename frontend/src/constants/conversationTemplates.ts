export interface ConversationTemplate {
  title: string;
  prompt: string;
  category: string;
}

export interface TemplateCategory {
  name: string;
  iconName: 'envelope' | 'academic-cap' | 'chat-bubble' | 'presentation';
}

export const TEMPLATE_CATEGORIES: TemplateCategory[] = [
  { name: 'メール添削', iconName: 'envelope' },
  { name: '敬語・表現', iconName: 'academic-cap' },
  { name: '報連相', iconName: 'chat-bubble' },
  { name: '会議・発表', iconName: 'presentation' },
];

export const CONVERSATION_TEMPLATES: ConversationTemplate[] = [
  { category: 'メール添削', title: 'メールの添削をお願い', prompt: '以下のメールをビジネスメールとして添削してください。改善点とその理由も教えてください。' },
  { category: 'メール添削', title: '件名の改善提案', prompt: 'メールの件名を効果的にするコツを教えてください。具体的な良い例・悪い例も挙げてください。' },
  { category: '敬語・表現', title: '敬語の使い分け', prompt: '尊敬語・謙譲語・丁寧語の使い分けについて、よくある間違いと正しい使い方を教えてください。' },
  { category: '敬語・表現', title: 'クッション言葉の練習', prompt: 'ビジネスシーンで使えるクッション言葉を状況別に教えてください。依頼・断り・質問それぞれの場面で使える表現をお願いします。' },
  { category: '報連相', title: '報告の構成方法', prompt: '上司への報告を分かりやすく伝えるための構成方法を教えてください。「結論→理由→詳細」の型で練習したいです。' },
  { category: '報連相', title: '悪い報告の伝え方', prompt: '問題やトラブルが発生した場合の報告の仕方を教えてください。相手を不安にさせずに正確に伝えるコツを知りたいです。' },
  { category: '会議・発表', title: '会議での発言の仕方', prompt: '会議で自分の意見を効果的に伝える方法を教えてください。発言のタイミングや話し方のコツもお願いします。' },
  { category: '会議・発表', title: '質疑応答の対処法', prompt: 'プレゼンテーション後の質疑応答で、うまく答えられない質問が来た場合の対処法を教えてください。' },
];
