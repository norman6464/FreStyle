import { useState } from 'react';
import { CodeBracketIcon } from '@heroicons/react/24/outline';

/**
 * LanguageIcon — 言語ロゴ（Devicon の SVG）を表示する。
 *
 * 言語ロゴは公式マークなので自作せず Devicon（MIT）の SVG を `public/lang/<key>.svg` に
 * vendoring して使う（出典・追加手順は public/lang/README.md）。key は backend の
 * `master_exercises.language` の値。
 *
 * 未知の言語やアイコン取得失敗時は Heroicons の汎用コードアイコンにフォールバックし、
 * 「アイコンが無いせいでカードが壊れて見える」状態を作らない。
 */
export default function LanguageIcon({
  language,
  className = 'w-10 h-10',
}: {
  language: string;
  className?: string;
}) {
  // 「失敗したかどうか」ではなく「どの言語で失敗したか」を持つ。 boolean だと、 言語一覧を
  // 行き来して language prop だけが変わったとき(コンポーネントは mount されたまま)に
  // 失敗状態が引き継がれ、 アイコンがある言語でもフォールバックのままになる。
  const [failedLanguage, setFailedLanguage] = useState<string | null>(null);

  if (failedLanguage === language) {
    return (
      <CodeBracketIcon
        className={`${className} text-[var(--color-text-muted)]`}
        aria-hidden="true"
      />
    );
  }

  return (
    <img
      src={`/lang/${language}.svg`}
      alt=""
      aria-hidden="true"
      loading="lazy"
      className={className}
      onError={() => setFailedLanguage(language)}
    />
  );
}
