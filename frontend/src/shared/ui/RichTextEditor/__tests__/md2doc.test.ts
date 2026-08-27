import { describe, it, expect } from 'vitest';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { getSchema } from '@tiptap/core';
import {
  markdownToDoc,
  restoreCodeBlockText,
  dedupeMarks,
  ensureListItemParagraph,
  assertDocMatchesSchema,
} from '../../../../../scripts/md2doc.mjs';
import { createSchemaExtensions } from '../schemaExtensions';
import { createEditorExtensions } from '../editorExtensions';

/** doc JSON ノードの最小形。変換器は素の JSON を返すのでテストもそれに合わせる。 */
interface DocNode {
  type: string;
  attrs?: Record<string, unknown>;
  content?: DocNode[];
  marks?: { type: string; attrs?: Record<string, unknown> }[];
  text?: string;
}

/** findNodes は doc を深さ優先で walk して type が一致するノードを集める。 */
function findNodes(doc: DocNode, type: string): DocNode[] {
  const found: DocNode[] = [];
  const walk = (node: DocNode) => {
    if (node.type === type) found.push(node);
    (node.content ?? []).forEach(walk);
  };
  walk(doc);
  return found;
}

/**
 * schemaSummary はスキーマの「意味のある部分」（ノード/マーク名・content 式・group・attrs 既定値・
 * excludes）だけを比較可能な形に落とす。toDOM 等の関数は同値比較できないため対象にしない。
 */
function schemaSummary(schema: ReturnType<typeof getSchema>) {
  return {
    nodes: Object.fromEntries(
      Object.entries(schema.nodes).map(([name, type]) => [
        name,
        {
          content: type.spec.content ?? null,
          group: type.spec.group ?? null,
          inline: type.spec.inline ?? false,
          attrs: Object.fromEntries(
            Object.entries(type.spec.attrs ?? {}).map(([attr, spec]) => [attr, spec.default]),
          ),
        },
      ]),
    ),
    marks: Object.fromEntries(
      Object.entries(schema.marks).map(([name, type]) => [
        name,
        { excludes: type.spec.excludes ?? null },
      ]),
    ),
  };
}

describe('createSchemaExtensions', () => {
  it('エディタ拡張と同一スキーマになる（NodeView / input rule の上掛けはスキーマを変えない）', () => {
    const editorSchema = getSchema(createEditorExtensions());
    const converterSchema = getSchema(createSchemaExtensions());
    expect(schemaSummary(editorSchema)).toEqual(schemaSummary(converterSchema));
  });

  it('ノード名 heading / codeBlock・マーク code（excludes 解除）を既存 doc と互換のまま持つ', () => {
    const schema = getSchema(createSchemaExtensions());
    for (const name of ['heading', 'codeBlock', 'table', 'taskList', 'taskItem', 'image']) {
      expect(schema.nodes[name], name).toBeDefined();
    }
    // CombinableCode: 排他を解いて太字等と併用できる（editorExtensions から移した仕様）。
    expect(schema.marks.code.spec.excludes).toBe('');
  });

  it('heading は levels 1〜3 のまま（エディタ UI と教材の章構造に合わせる）', () => {
    const heading = createSchemaExtensions().find((extension) => extension.name === 'heading');
    expect(heading?.options.levels).toEqual([1, 2, 3]);
  });

  it('image: false でスキーマから画像ノードを外せる（エディタの既存オプションと同じ挙動）', () => {
    const schema = getSchema(createSchemaExtensions({ image: false }));
    expect(schema.nodes.image).toBeUndefined();
  });
});

describe('markdownToDoc', () => {
  it('見出しと本文を変換する', () => {
    const doc = markdownToDoc('# タイトル\n\n## 節\n\n本文。\n');
    expect(doc.content[0]).toEqual({
      type: 'heading',
      attrs: { level: 1 },
      content: [{ type: 'text', text: 'タイトル' }],
    });
    expect(doc.content[1].attrs).toEqual({ level: 2 });
    expect(doc.content[2].type).toBe('paragraph');
  });

  it('GFM 表を table / tableRow / tableHeader / tableCell へ変換する', () => {
    const doc = markdownToDoc('| 列A | 列B |\n| --- | --- |\n| 1 | 2 |\n');
    const [table] = findNodes(doc, 'table');
    expect(table).toBeDefined();
    const rows = findNodes(table, 'tableRow');
    expect(rows).toHaveLength(2);
    expect(findNodes(rows[0], 'tableHeader')).toHaveLength(2);
    expect(findNodes(rows[1], 'tableCell')).toHaveLength(2);
  });

  it('タスクリストを taskList / taskItem（checked 属性つき）へ変換する', () => {
    const doc = markdownToDoc('- [ ] 未完了\n- [x] 完了\n');
    expect(findNodes(doc, 'taskList')).toHaveLength(1);
    const items = findNodes(doc, 'taskItem');
    expect(items.map((item) => item.attrs?.checked)).toEqual([false, true]);
  });

  it('言語付きコードブロックを language 属性つきで原文どおり変換する', () => {
    const doc = markdownToDoc('```go\nfunc main() {\n\tprintln("hi")\n}\n```\n');
    const [block] = findNodes(doc, 'codeBlock');
    expect(block.attrs?.language).toBe('go');
    expect(block.content?.[0].text).toBe('func main() {\n\tprintln("hi")\n}');
  });

  it('リスト内フェンスはリストのインデントだけを剥がし、意図的な先頭スペースを保つ', () => {
    const md = '- 手順:\n\n  ```sql\n  EXPLAIN\n    SELECT 1;\n  ```\n';
    const doc = markdownToDoc(md);
    const [block] = findNodes(doc, 'codeBlock');
    expect(block.attrs?.language).toBe('sql');
    // リスト由来の 2 スペースは除去し、SELECT 前の 2 スペース（原文の整形）は残ること。
    expect(block.content?.[0].text).toBe('EXPLAIN\n  SELECT 1;');
  });

  it('画像を image ノード（src 属性）へ変換する', () => {
    const doc = markdownToDoc('![図](https://example.com/x.png)\n');
    const [image] = findNodes(doc, 'image');
    expect(image.attrs?.src).toBe('https://example.com/x.png');
  });

  it('Markdown のリンクを link マーク（href）へ変換する', () => {
    const doc = markdownToDoc('[公式サイト](https://example.com/a)\n');
    const [paragraph] = findNodes(doc, 'paragraph');
    expect(paragraph.content?.[0].text).toBe('公式サイト');
    expect(paragraph.content?.[0].marks).toEqual([
      { type: 'link', attrs: { href: 'https://example.com/a', title: null } },
    ]);
  });

  it('許可できないスキームのリンクはマークを外して文字だけ残す（教材からも危険な href を入れない）', () => {
    const doc = markdownToDoc('[押して](javascript:alert(1))\n');
    const [paragraph] = findNodes(doc, 'paragraph');
    expect(paragraph.content?.[0].text).toBe('押して');
    expect(paragraph.content?.[0].marks).toBeUndefined();
  });

  it('変換結果はアプリスキーマの Node.check() を通る', () => {
    const md = [
      '# 章',
      '',
      '> 引用と **太字** と `code`',
      '',
      '1. 手順1',
      '2. 手順2',
      '   - 入れ子',
      '',
      '```bash',
      'echo hello',
      '```',
      '',
      '- [ ] TODO',
      '',
      '| a | b |',
      '| - | - |',
      '| 1 | 2 |',
      '',
    ].join('\n');
    const doc = markdownToDoc(md);
    expect(() => assertDocMatchesSchema(doc)).not.toThrow();
  });
});

describe('dedupeMarks', () => {
  it('同種マークの重複を除去し、attrs が異なる同種マークは残す', () => {
    const doc: DocNode = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: '強調',
              marks: [
                { type: 'bold' },
                { type: 'bold' },
                { type: 'link', attrs: { href: 'https://a.example' } },
                { type: 'link', attrs: { href: 'https://b.example' } },
              ],
            },
          ],
        },
      ],
    };
    dedupeMarks(doc);
    expect(doc.content?.[0].content?.[0].marks).toEqual([
      { type: 'bold' },
      { type: 'link', attrs: { href: 'https://a.example' } },
      { type: 'link', attrs: { href: 'https://b.example' } },
    ]);
  });
});

describe('ensureListItemParagraph', () => {
  it('先頭が block で始まる listItem / taskItem に空 paragraph を補う', () => {
    const doc: DocNode = {
      type: 'doc',
      content: [
        {
          type: 'bulletList',
          content: [
            {
              type: 'listItem',
              content: [{ type: 'bulletList', content: [] }],
            },
          ],
        },
        {
          type: 'taskList',
          content: [
            {
              type: 'taskItem',
              attrs: { checked: false },
              content: [{ type: 'paragraph', content: [{ type: 'text', text: 'ok' }] }],
            },
          ],
        },
      ],
    };
    ensureListItemParagraph(doc);
    // 先頭が入れ子リストだった listItem には paragraph が補われる。
    expect(doc.content?.[0].content?.[0].content?.[0].type).toBe('paragraph');
    // 先頭が paragraph の taskItem はそのまま。
    expect(doc.content?.[1].content?.[0].content).toHaveLength(1);
  });
});

describe('restoreCodeBlockText', () => {
  it('フェンス数と codeBlock 数が一致しないときは throw する（黙って取り違えない）', () => {
    const doc: DocNode = { type: 'doc', content: [{ type: 'paragraph' }] };
    expect(() => restoreCodeBlockText('```\nx\n```\n', doc)).toThrow(/一致しない/);
  });

  it('チルダフェンス（~~~）も 1 ブロックとして原文どおり書き戻す', () => {
    const doc = markdownToDoc('~~~\n  indented body\n~~~\n');
    const blocks = findNodes(doc, 'codeBlock');
    expect(blocks).toHaveLength(1);
    expect(blocks[0].content?.[0].text).toBe('  indented body');
  });

  it('4 連バッククォートの中の ``` は閉じと誤認しない（同文字・同長以上のみ閉じ）', () => {
    const doc = markdownToDoc('````md\n```\ninner fence\n```\n````\n');
    const blocks = findNodes(doc, 'codeBlock');
    expect(blocks).toHaveLength(1);
    expect(blocks[0].content?.[0].text).toBe('```\ninner fence\n```');
  });

  it('閉じフェンスは開始より長くてもよい（``` を ````` で閉じる）', () => {
    const doc = markdownToDoc('```\nbody\n`````\n');
    const blocks = findNodes(doc, 'codeBlock');
    expect(blocks).toHaveLength(1);
    expect(blocks[0].content?.[0].text).toBe('body');
  });

  it('インデント形式のコードブロックは書き戻し不能として変換を拒否する', () => {
    // 空行 + 4 スペースはパーサが codeBlock 化するが原文にフェンスが無い → 件数不一致で throw。
    expect(() => markdownToDoc('段落\n\n    indented code\n')).toThrow(/フェンス/);
  });
});


describe('md2doc CLI の実行契約', () => {
  const cliPath = resolve(process.cwd(), 'scripts/md2doc.mjs');
  const tsxBin = resolve(process.cwd(), 'node_modules/.bin/tsx');

  /** runCli は CLI を子プロセスで実行し、stdout / stderr / exit code を返す。 */
  function runCli(args: string[], stdin?: string) {
    try {
      const stdout = execFileSync(tsxBin, [cliPath, ...args], {
        input: stdin,
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'pipe'],
      });
      return { stdout, stderr: '', status: 0 };
    } catch (error) {
      const failed = error as { stdout?: string; stderr?: string; status?: number };
      return { stdout: failed.stdout ?? '', stderr: failed.stderr ?? '', status: failed.status ?? 1 };
    }
  }

  it('単一ファイル: doc JSON を stdout へ出す / batch 失敗時は stdout を出さず exit 1', () => {
    const dir = mkdtempSync(join(tmpdir(), 'md2doc-'));
    try {
      const ok = join(dir, 'ok.md');
      writeFileSync(ok, '# 見出し\n\n本文。\n');

      // 単一ファイル: 正常。
      const single = runCli([ok]);
      expect(single.status).toBe(0);
      expect(JSON.parse(single.stdout).type).toBe('doc');

      // batch: 正常（__proto__ を含むパス名でも own property として出力される）。
      const proto = join(dir, '__proto__');
      writeFileSync(proto, 'proto という名前のファイル\n');
      const batch = runCli(['--batch'], JSON.stringify([ok, proto]));
      expect(batch.status).toBe(0);
      const map = JSON.parse(batch.stdout) as Record<string, { type: string }>;
      expect(Object.keys(map)).toHaveLength(2);
      expect(map[proto].type).toBe('doc');

      // batch: 1 件でも失敗したら map を出さず exit 1 + stderr に対象と理由。
      const missing = join(dir, 'missing.md');
      const failed = runCli(['--batch'], JSON.stringify([ok, missing]));
      expect(failed.status).toBe(1);
      expect(failed.stdout).toBe('');
      expect(failed.stderr).toContain('missing.md');

      // stdin が JSON 配列でないときも exit 1。
      const badInput = runCli(['--batch'], '{"not":"array"}');
      expect(badInput.status).toBe(1);

      // 引数なしは Usage + exit 1。
      const usage = runCli([]);
      expect(usage.status).toBe(1);
      expect(usage.stderr).toContain('Usage');
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  }, 60000);
});
