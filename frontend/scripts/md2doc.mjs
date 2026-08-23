/**
 * md2doc.mjs — 教材 Markdown → tiptap(ProseMirror) doc JSON 変換器。
 *
 * 教材リポの seed パイプラインが md から tiptap doc(jsonb) を生成して DB へ投入するために使う。
 * スキーマはアプリのエディタと同じ createSchemaExtensions()（単一ソース）から組み立てるので、
 * 変換結果は必ずエディタで開ける形になる（独自の拡張配列を持たない＝二重管理しない）。
 *
 * 使い方:
 *   node scripts/md2doc.mjs <file.md>             # doc JSON を stdout へ
 *   node scripts/md2doc.mjs --batch < paths.json  # stdin の JSON 配列 ["a.md", ...] を
 *                                                 # {"a.md": {doc}, ...} の JSON map で stdout へ
 * 変換エラー時は対象ファイルと理由を stderr に出して exit 1（batch は 1 件でも失敗したら
 * map を出力しない — 部分成功の doc で DB を汚さないため）。
 *
 * schemaExtensions.ts は TypeScript のまま import する（Node 22.18+ の型ストリップが既定で
 * 有効なため。erasable な構文のみで書かれている）。
 */
import { readFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';
import { getSchema, resolveExtensions } from '@tiptap/core';
import { Markdown, MarkdownManager } from '@tiptap/markdown';
import { createSchemaExtensions } from '../src/shared/ui/RichTextEditor/schemaExtensions.ts';

// MarkdownManager はフラット化済みの拡張配列を受け取る（editor 不要・Node で動く）。
function makeManager() {
  const flat = resolveExtensions([...createSchemaExtensions(), Markdown]);
  return new MarkdownManager({ extensions: flat });
}

let cachedSchema = null;

/**
 * assertDocMatchesSchema は doc がアプリのエディタスキーマで有効かを検証する（不正なら throw）。
 * DB へ入れた doc がエディタで開けない事故を、変換時点で止めるための最終ゲート。
 */
export function assertDocMatchesSchema(doc) {
  cachedSchema ??= getSchema(createSchemaExtensions());
  cachedSchema.nodeFromJSON(doc).check();
}

/** markdownToDoc は md 原文をアプリスキーマの doc JSON へ変換し、スキーマ検証まで行う。 */
export function markdownToDoc(markdown) {
  const manager = makeManager();
  const node = manager.parse(markdown);
  const doc = typeof node.toJSON === 'function' ? node.toJSON() : node;
  restoreCodeBlockText(markdown, doc);
  dedupeMarks(doc);
  ensureListItemParagraph(doc);
  assertDocMatchesSchema(doc);
  return doc;
}

/**
 * ensureListItemParagraph は listItem / taskItem の先頭が block（入れ子リスト等）で始まる場合に
 * 空の paragraph を先頭へ補う。tiptap のスキーマは listItem content を 'paragraph block*' と
 * 定義しており、親テキストの無いぶら下げリスト（md の `- + 子` のような記法）が invalid になるため。
 */
export function ensureListItemParagraph(doc) {
  const walk = (n) => {
    if ((n.type === 'listItem' || n.type === 'taskItem') && Array.isArray(n.content)) {
      if (n.content.length > 0 && n.content[0].type !== 'paragraph') {
        n.content.unshift({ type: 'paragraph' });
      }
    }
    (n.content ?? []).forEach(walk);
  };
  walk(doc);
}

/**
 * dedupeMarks は text ノードの同種マーク重複（bold,bold / italic,italic 等）を除去する。
 * 原文の入れ子強調（** の中の **）を markdown パーサが重複マークとして平坦化することがあり、
 * ProseMirror のスキーマ検証（Node.check）は同種マークの重複を許さないため正規化する。
 * 表示は同種マーク 1 つと等価（欠落なし）。attrs が異なる同種マークは別物なので残す。
 */
export function dedupeMarks(doc) {
  const walk = (n) => {
    if (n.type === 'text' && Array.isArray(n.marks) && n.marks.length > 1) {
      const seen = new Set();
      n.marks = n.marks.filter((m) => {
        const key = JSON.stringify([m.type, m.attrs ?? null]);
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      });
    }
    (n.content ?? []).forEach(walk);
  };
  walk(doc);
}

/**
 * restoreCodeBlockText は doc 内の codeBlock 本文を md 原文のフェンス内容で置き換える。
 * markdown パーサはリスト内フェンスのインデント除去が CommonMark と 1 文字ずれることがあり
 * （余分な先頭スペースが残る）、逆に EXPLAIN 出力等の意図的な先頭スペースは保持すべきで、
 * 後段のヒューリスティック除去では区別できない。フェンスと codeBlock は 1:1 で順序保存
 * されるため、原文をそのまま書き戻すのが最も安全。
 */
export function restoreCodeBlockText(markdown, doc) {
  const fences = [];
  const lines = markdown.split('\n');
  let i = 0;
  while (i < lines.length) {
    const open = /^(\s*)(> ?)?(`{3}|｀{3})(.*)$/.exec(lines[i]);
    if (!open) { i++; continue; }
    const indent = open[1].length;
    const quoted = Boolean(open[2]);
    const body = [];
    i++;
    while (i < lines.length && !/^\s*(> ?)?(`{3}|｀{3})\s*$/.test(lines[i])) {
      let line = lines[i];
      if (quoted) line = line.replace(/^\s*> ?/, '');
      else if (indent > 0) {
        let strip = 0;
        while (strip < indent && line[strip] === ' ') strip++;
        line = line.slice(strip);
      }
      body.push(line);
      i++;
    }
    i++; // 閉じフェンスを飛ばす
    fences.push(body.join('\n'));
  }

  const blocks = [];
  const walk = (n) => {
    if (n.type === 'codeBlock') blocks.push(n);
    (n.content ?? []).forEach(walk);
  };
  walk(doc);
  if (blocks.length !== fences.length) {
    throw new Error(`codeBlock 数 (${blocks.length}) とフェンス数 (${fences.length}) が一致しない`);
  }
  blocks.forEach((block, idx) => {
    const text = fences[idx];
    block.content = text === '' ? [] : [{ type: 'text', text }];
  });
}

function convertFile(path) {
  return markdownToDoc(readFileSync(path, 'utf8'));
}

function reason(error) {
  return error instanceof Error ? error.message : String(error);
}

function runBatch() {
  const paths = JSON.parse(readFileSync(0, 'utf8')); // fd 0 = stdin
  if (!Array.isArray(paths) || paths.some((p) => typeof p !== 'string')) {
    process.stderr.write('stdin は ["path.md", ...] の JSON 配列であること\n');
    process.exit(1);
  }
  const result = {};
  const errors = [];
  for (const path of paths) {
    try {
      result[path] = convertFile(path);
    } catch (error) {
      errors.push(`${path}: ${reason(error)}`);
    }
  }
  if (errors.length > 0) {
    for (const line of errors) process.stderr.write(`${line}\n`);
    process.exit(1);
  }
  process.stdout.write(`${JSON.stringify(result)}\n`);
}

function runSingle(path) {
  try {
    process.stdout.write(`${JSON.stringify(convertFile(path))}\n`);
  } catch (error) {
    process.stderr.write(`${path}: ${reason(error)}\n`);
    process.exit(1);
  }
}

// vitest 等から関数 import されたときに CLI が走らないよう、直接実行のときだけ main を動かす。
const isDirectRun =
  process.argv[1] !== undefined && import.meta.url === pathToFileURL(process.argv[1]).href;
if (isDirectRun) {
  const arg = process.argv[2];
  if (arg === '--batch') {
    runBatch();
  } else if (arg) {
    runSingle(arg);
  } else {
    process.stderr.write(
      'Usage: node scripts/md2doc.mjs <file.md>\n' +
        '       node scripts/md2doc.mjs --batch < paths.json\n',
    );
    process.exit(1);
  }
}
