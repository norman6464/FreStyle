/**
 * 実行エラー(stderr)から「何行目のエラーか」を言語別に抽出する(FRESTYLE-117)。
 * 実行系(backend sandbox)は内部パスを ./main.go 等に整形して返すため、
 * stderr の行番号はエディタの行番号とそのまま一致する。
 */
export interface ExecutionErrorMarker {
  /** 1 始まりの行番号。 */
  line: number;
  /** その行に紐づくエラーメッセージ(ホバー表示用)。 */
  message: string;
}

// 言語ごとの「行番号つきエラー行」のパターン。1 番目のキャプチャが行番号。
const LINE_PATTERNS: Record<string, RegExp> = {
  go: /main\.go:(\d+)(?::\d+)?:/,
  javascript: /main\.js:(\d+)(?::\d+)?/,
  typescript: /main\.ts:(\d+)(?::\d+)?/,
  php: /on line (\d+)/,
  bash: /script\.sh: (?:line )?(\d+):/,
};

/**
 * stderr を行単位に走査し、行番号を含む行をマーカーに変換する。
 * 同じ行番号が複数回出る場合は最初のメッセージを採用する。
 */
export function parseErrorLines(stderr: string, language: string): ExecutionErrorMarker[] {
  const pattern = LINE_PATTERNS[language.toLowerCase()];
  if (!pattern || !stderr) return [];

  const byLine = new Map<number, string>();
  for (const raw of stderr.split('\n')) {
    const m = raw.match(pattern);
    if (!m) continue;
    const line = Number(m[1]);
    if (!Number.isInteger(line) || line < 1) continue;
    if (!byLine.has(line)) {
      byLine.set(line, raw.trim());
    }
  }

  // node のスタックトレースのように「行番号の行」と「例外メッセージの行」が分かれる形式では、
  // 例外メッセージ(例: SyntaxError: ...)をホバーに添える方が原因が分かりやすい。
  const exception = stderr.split('\n').find((l) => /^[A-Za-z]*(?:Error|Exception): /.test(l.trim()));
  const result = [...byLine.entries()]
    .sort((a, b) => a[0] - b[0])
    .map(([line, message]) => ({
      line,
      message: exception && !message.includes(exception.trim()) ? `${message}\n${exception.trim()}` : message,
    }));
  return result;
}
