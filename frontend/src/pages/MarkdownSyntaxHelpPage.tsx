import { Link } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';

/**
 * MarkdownSyntaxHelpPage — `/notes/markdown-help` で表示される Markdown 記法のチートシート。
 *
 * 構成:
 *   - 左カラムに記法（コードブロック）/ 右カラムにレンダリング結果
 *   - ノートの編集中に並べて参照できるように、 同じ全幅レイアウト
 *
 * Markdown は GitHub Flavored Markdown (remark-gfm) + remark の拡張記法をサポート。
 * 表示はノート / 教材ページの NoteMarkdownEditor の Preview と同じレンダラーで動く前提。
 */
export default function MarkdownSyntaxHelpPage() {
  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-5xl mx-auto space-y-6">
      <header className="space-y-2">
        <Link
          to="/notes"
          className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
        >
          <ArrowLeftIcon className="w-3.5 h-3.5" />
          ノートに戻る
        </Link>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Markdown 記法ヘルプ</h1>
        <p className="text-sm text-[var(--color-text-tertiary)]">
          ノート / 教材で使える Markdown の主要な記法をまとめたチートシートです。
          コピーしてエディタに貼り付けるとそのまま動きます。
        </p>
      </header>

      <Section title="見出し">
        <Row code={`# 見出し 1\n## 見出し 2\n### 見出し 3`}>
          <h1 className="text-2xl font-bold m-0 text-[var(--color-text-primary)]">見出し 1</h1>
          <h2 className="text-xl font-semibold mt-2 text-[var(--color-text-primary)]">見出し 2</h2>
          <h3 className="text-base font-semibold mt-2 text-[var(--color-text-primary)]">見出し 3</h3>
        </Row>
      </Section>

      <Section title="文字装飾">
        <Row code={`**太字** *斜体* ~~取り消し線~~\n\`インラインコード\``}>
          <p className="m-0 text-[var(--color-text-primary)]">
            <strong>太字</strong> <em>斜体</em> <s>取り消し線</s>{' '}
            <code className="px-1 py-0.5 rounded bg-[var(--color-surface-3)] text-[0.85em]">
              インラインコード
            </code>
          </p>
        </Row>
      </Section>

      <Section title="リスト">
        <Row
          code={`- 順序なし 1\n- 順序なし 2\n  - ネスト\n\n1. 順序あり 1\n2. 順序あり 2`}
        >
          <ul className="m-0 list-disc pl-5 text-[var(--color-text-primary)]">
            <li>順序なし 1</li>
            <li>
              順序なし 2
              <ul className="list-disc pl-5">
                <li>ネスト</li>
              </ul>
            </li>
          </ul>
          <ol className="mt-2 m-0 list-decimal pl-5 text-[var(--color-text-primary)]">
            <li>順序あり 1</li>
            <li>順序あり 2</li>
          </ol>
        </Row>
      </Section>

      <Section title="チェックリスト (GFM)">
        <Row code={`- [ ] 未完了\n- [x] 完了`}>
          <ul className="m-0 list-none pl-0 text-[var(--color-text-primary)]">
            <li className="flex items-center gap-2">
              <input type="checkbox" disabled />
              未完了
            </li>
            <li className="flex items-center gap-2">
              <input type="checkbox" disabled checked readOnly />
              完了
            </li>
          </ul>
        </Row>
      </Section>

      <Section title="リンク / 画像">
        <Row
          code={`[FreStyle](https://normanblog.com)\n\n![代替テキスト](https://placehold.jp/120x40.png)`}
        >
          <p className="m-0">
            <a
              href="https://normanblog.com"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary-400 underline-offset-2 hover:underline"
            >
              FreStyle
            </a>
          </p>
          <img
            src="https://placehold.jp/120x40.png"
            alt="代替テキスト"
            className="mt-2 rounded"
          />
        </Row>
      </Section>

      <Section title="引用">
        <Row code={`> 引用ブロック\n> 改行も入れられる`}>
          <blockquote className="m-0 border-l-2 border-surface-3 pl-3 text-[var(--color-text-secondary)] italic">
            引用ブロック<br />
            改行も入れられる
          </blockquote>
        </Row>
      </Section>

      <Section title="コードブロック (言語指定)">
        <Row
          code={'```js\nfunction hello(name) {\n  return `Hello, ${name}!`;\n}\n```'}
        >
          <pre className="m-0 p-2 rounded bg-[var(--color-surface-3)] overflow-x-auto text-xs">
            <code>
              {`function hello(name) {\n  return \`Hello, \${name}!\`;\n}`}
            </code>
          </pre>
        </Row>
      </Section>

      <Section title="表 (GFM)">
        <Row
          code={`| 列1 | 列2 | 列3 |\n|---|---|---|\n| a | b | c |\n| d | e | f |`}
        >
          <table className="m-0 text-xs border-collapse">
            <thead>
              <tr>
                <th className="border border-surface-3 px-2 py-1">列1</th>
                <th className="border border-surface-3 px-2 py-1">列2</th>
                <th className="border border-surface-3 px-2 py-1">列3</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="border border-surface-3 px-2 py-1">a</td>
                <td className="border border-surface-3 px-2 py-1">b</td>
                <td className="border border-surface-3 px-2 py-1">c</td>
              </tr>
              <tr>
                <td className="border border-surface-3 px-2 py-1">d</td>
                <td className="border border-surface-3 px-2 py-1">e</td>
                <td className="border border-surface-3 px-2 py-1">f</td>
              </tr>
            </tbody>
          </table>
        </Row>
      </Section>

      <Section title="水平線">
        <Row code={`---`}>
          <hr className="m-0 border-surface-3" />
        </Row>
      </Section>

      <Section title="自動リンク (GFM)">
        <Row code={`https://normanblog.com を直接書くだけでリンクになります`}>
          <p className="m-0 text-[var(--color-text-primary)]">
            <a
              href="https://normanblog.com"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary-400 underline-offset-2 hover:underline"
            >
              https://normanblog.com
            </a>
            を直接書くだけでリンクになります
          </p>
        </Row>
      </Section>

      <div className="rounded-lg border border-blue-500/30 bg-blue-500/5 p-4 text-sm text-[var(--color-text-secondary)]">
        💡 もっと知りたい場合は{' '}
        <a
          href="https://github.github.com/gfm/"
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary-400 underline-offset-2 hover:underline"
        >
          GitHub Flavored Markdown 仕様
        </a>{' '}
        を参照してください。
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="space-y-2">
      <h2 className="text-sm font-semibold text-[var(--color-text-secondary)] tracking-wide">
        {title}
      </h2>
      {children}
    </section>
  );
}

function Row({ code, children }: { code: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      <CodeSnippet code={code} />
      <div className="rounded-md border border-surface-3 bg-surface-1 p-3 prose prose-sm max-w-none">
        {children}
      </div>
    </div>
  );
}

function CodeSnippet({ code }: { code: string }) {
  return (
    <div className="relative">
      <pre className="m-0 p-3 rounded-md bg-surface-2 border border-surface-3 text-xs font-mono text-[var(--color-text-primary)] whitespace-pre-wrap break-words">
        {code}
      </pre>
      <CopyButton text={code} />
    </div>
  );
}

import { useState } from 'react';
import { ClipboardDocumentIcon, CheckIcon } from '@heroicons/react/24/outline';

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch {
      // クリップボード非対応環境では sileng fail。
    }
  };
  return (
    <button
      type="button"
      onClick={handleCopy}
      aria-label="コードをコピー"
      className="absolute top-2 right-2 p-1.5 rounded bg-surface-3/60 hover:bg-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] transition-colors"
    >
      {copied ? (
        <CheckIcon className="w-3.5 h-3.5 text-green-400" />
      ) : (
        <ClipboardDocumentIcon className="w-3.5 h-3.5" />
      )}
    </button>
  );
}
