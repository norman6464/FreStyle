import { Link } from 'react-router-dom';
import {
  HomeIcon,
  SparklesIcon,
  CodeBracketIcon,
  DocumentTextIcon,
  DocumentChartBarIcon,
} from '@heroicons/react/24/outline';

/**
 * シンプルなホーム画面。
 * 旧版で並んでいたゲーミフィケーション系（バッジ / レベル / チャレンジ / 名言 等）は
 * すべて削除し、コア機能（AI / コード学習 / ノート / レポート）への導線だけを残す。
 *
 * 設計判断:
 *   - ヘッダーのアイコン + タイトルで現在地を明示
 *   - 4 枚のリンクカードで主要機能に直接遷移
 *   - 表示する数値・進捗系は一切持たない（store / API 依存なし）
 */
export default function MenuPage() {
  return (
    <div className="px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <header className="flex items-center gap-2">
        <HomeIcon className="w-6 h-6 text-[var(--color-text-muted)]" />
        <h1 className="text-2xl font-bold">ホーム</h1>
      </header>

      <p className="text-sm text-[var(--color-text-muted)] leading-relaxed">
        左のメニューから機能を選んでください。下のカードからもアクセスできます。
      </p>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <NavCard
          to="/chat/ask-ai"
          icon={SparklesIcon}
          title="AI チャット"
          description="質問・要約・コード補助など、自由に対話できます。"
        />
        <NavCard
          to="/code-editor"
          icon={CodeBracketIcon}
          title="コード学習"
          description="演習問題を解いて手を動かしながら学べます。"
        />
        <NavCard
          to="/notes"
          icon={DocumentTextIcon}
          title="ノート"
          description="学習メモを残す・振り返ることができます。"
        />
        <NavCard
          to="/reports"
          icon={DocumentChartBarIcon}
          title="レポート"
          description="月次の学習サマリーを確認できます。"
        />
      </div>
    </div>
  );
}

interface NavCardProps {
  to: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  title: string;
  description: string;
}

function NavCard({ to, icon: Icon, title, description }: NavCardProps) {
  return (
    <Link
      to={to}
      className="block p-4 rounded-lg border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] hover:bg-[var(--color-surface-2)] transition-colors"
    >
      <div className="flex items-start gap-3">
        <Icon className="w-5 h-5 mt-0.5 text-[var(--color-text-muted)] flex-shrink-0" />
        <div className="flex-1">
          <h2 className="font-medium text-[var(--color-text-primary)]">{title}</h2>
          <p className="text-xs text-[var(--color-text-muted)] mt-1">{description}</p>
        </div>
      </div>
    </Link>
  );
}
