# AI チャットのストリーミング表示と語単位フェードイン

`/ask-ai`（[AskAiPage](../frontend/src/pages/AskAiPage.tsx)）のアシスタント応答を SSE で受けて描画する仕組みと、
本文を Gemini 風に「新しく現れた語がふわっとフェードインする」表示にする実装（FRESTYLE-145）について。

## ストリーミングの描画経路

1. 受信: [useAiChatSse](../frontend/src/hooks/useAiChatSse.ts) が `fetch` + `ReadableStream` で SSE を読み、
   `event: token` / `data: {"delta": "..."}` をパースして `onEvent` に渡す（標準 `EventSource` は POST 不可のため使わない）。
2. 反映: [useAskAi](../frontend/src/hooks/useAskAi.ts) が送信時に `id = "streaming-{ts}"`・`content: ""` の
   アシスタント placeholder を 1 つ積み、`token` ごとに `content` の末尾へ `delta` を**そのまま文字連結**する。
   `done` で `id` を backend の確定 id に、`content` を全文に差し替える。
3. 描画: [AskAiPage](../frontend/src/pages/AskAiPage.tsx) の `messages.map` が `key={message.id}` で
   [MessageBubble](../frontend/src/components/MessageBubble.tsx) を並べる。本文は
   [MarkdownView](../frontend/src/components/message/MarkdownView.tsx)（react-markdown）で描画され、
   **token 追記のたびに `content` 文字列全体が再パースされる**。

`content` が空の間は「考え中...」（favicon の `animate-thinking` パルス）を出し、最初の token で本文へ切り替わる。

## 語単位フェードイン（FRESTYLE-145）

### 何を真似たか（Gemini 調査結果）

Gemini の本文ストリーミングは 2 層構造:

- **バッファ＆リペース**: モデルは大小まちまちの大きい塊で返すため、UI 側で溜めて一定レートで放出する。
- **語（word）単位のフェードイン**: 放出された各セグメントを `opacity: 0→1` を主役に、可読性のため控えめな
  `blur(数px→0)` を併用（~200〜400ms・ease-out）。各セグメントは **mount 時に一度だけ**再生し、既表示テキストは
  再アニメしない。日本語など空白で区切らない言語は `Intl.Segmenter`（`granularity: 'word'`）で語分節する。

（※よく見つかる「1500ms・`background-position-x` のグラデーション横ワイプ」は Gemini の**トップ挨拶文の別アニメ**で、
チャット本文のフェードとは別系統。）

### 実装方式

react-markdown の **rehype 段でテキストノードを語 `<span class="fade-seg">` に分割**し、CSS でフェードさせる
（[rehypeFadeSegments](../frontend/src/components/message/rehypeFadeSegments.ts)）。

- **ストリーミング中だけ**適用する（`MarkdownView` の `isStreaming` prop。判定は `String(id).startsWith('streaming-')`
  ＝末尾 bookend favicon の出し分けと同じ signal）。完了後・履歴・テストでは**素の Markdown（span なし）**に戻すので、
  DOM が軽く保たれ、既存テストの `getByText` 完全一致も維持される。
- 語分割は `Intl.Segmenter('ja', {granularity:'word'})`（無い環境は空白区切りにフォールバック）。**空白のみの
  セグメントは span で包まず素のテキストで残す**（折り返し・レイアウトを保つ）。
- **`pre` / `code` 配下は分割しない**（`rehype-highlight` のハイライトとコピー UI を壊さない／コードを 1 文字ずつ
  光らせない）。`rehypePlugins` は `[rehypeHighlight, rehypeFadeSegments]` の順（highlight 済みの code は skip される）。

### 「既表示テキストを再アニメさせない」仕組み（最重要）

再パースのたびに全 hast を作り直すが、**追記のみのストリーミングでは既出プレフィックスの語 span は位置（順序）が
不変**なので、React が DOM ノードを再利用する（＝CSS アニメが再生し直されない）。新しく mount した末尾の語 span
だけが 1 回フェードする。この「プレフィックスの span DOM が同一ノードとして再利用される」ことは
[MarkdownView.test](../frontend/src/components/message/__tests__/MarkdownView.test.tsx) で **DOM ノードの同一性**を
アサートして守っている（崩れると全文チラつきになるため）。

char-offset で「既出語」を潰して `animation: none` にする方式は採らなかった。フェード途中の語が次 token で
「既出」判定され opacity 1 へ瞬間スナップする（pop）副作用があるため、DOM 再利用に委ねる方が滑らか。

なお、**末尾ブロックの種別が確定する瞬間**（段落→見出し/表、`**強調**` が閉じて `<strong>` になる等）だけは、
その部分の subtree が unmount→mount され一度だけ再フェードする。影響は「今まさにストリーミング中の末尾」に限定され、
確定済みの先行テキストは再フェードしないため、全文チラつきにはならない（許容する。char-offset skip の pop を避けた
トレードオフ）。

### CSS とアクセシビリティ

[index.css](../frontend/src/index.css) に `@keyframes fade-seg-in` と `.fade-seg` を定義（`thinking-pulse` と同じ流儀）。

- 主役は **opacity**。`.fade-seg` は inline 要素なので `transform`（translateY）は効かない／使わない。高さも変えないので
  ストリーミング中の**自動スクロール追従**（`distanceFromBottom` 判定）を乱さない。
- **blur は `prefers-reduced-motion: no-preference` の時だけ**上乗せ（`@keyframes` を media 内で再定義。
  `4px→0` の 1 回きりで、毎フレーム連続 animate はしない）。GPU コストの高い blur を「動き OK」ユーザーに限定する。
- **`prefers-reduced-motion: reduce`** では `animation-duration: 0.01ms`（実質即時。`animationend` を残すため 0 にしない）。

## テスト

- [rehypeFadeSegments.test](../frontend/src/components/message/__tests__/rehypeFadeSegments.test.ts): 語分割・テキスト保持・
  `pre`/`code` 非分割・空白の扱い。
- [MarkdownView.test](../frontend/src/components/message/__tests__/MarkdownView.test.tsx): `isStreaming` の有無での分割/非分割、
  コード非分割、**DOM ノード再利用（再フェード防止）**、見出し配下、GFM（リンク/表/太字）との共存。
- [MessageBubble.test](../frontend/src/components/__tests__/MessageBubble.test.tsx): `id` が `streaming-` 始まりのとき
  fade span・bookend 非表示／確定 id では素の描画・bookend 表示。

## 今後の選択肢（未実装）

実機で「高速モデルが 1 commit に大量語を返してもたつく」場合は、`useAskAi` の token 反映に軽い
reveal バッファ（`requestAnimationFrame` で 1 フレーム 1 フラッシュ）を足して放出レートを分離する余地がある
（Gemini のバッファ＆リペースに相当）。長文で jank が出る場合は、段落ブロック単位の `React.memo`（raw ソースキー）で
末尾以外の再パースを止める最適化も可能。いずれも今回はスコープ外。
