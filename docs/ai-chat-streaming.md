# AI チャットのストリーミング表示（Gemini 実物仕様のペーシング＋フェードイン）

`/chat/ask-ai`（[AskAiPage](../frontend/src/pages/AskAiPage.tsx)）のアシスタント応答を SSE で受けて描画する仕組みと、
本文を Gemini と同じリズム・アニメーションで表示する実装（FRESTYLE-145 → 146）について。

## ストリーミングの描画経路

1. 受信: [useAiChatSse](../frontend/src/hooks/useAiChatSse.ts) が `fetch` + `ReadableStream` で SSE を読み、
   `event: token` / `data: {"delta": "..."}` をパースして `onEvent` に渡す（標準 `EventSource` は POST 不可のため使わない）。
2. 反映: [useAskAi](../frontend/src/hooks/useAskAi.ts) が送信時に `id = clientId = "streaming-{ts}"`・`content: ""` の
   アシスタント placeholder を積み、`token` ごとに `content` へ `delta` を文字連結する。`done` で `id` を backend の
   確定 id に、`content` を全文に差し替える（`clientId` は spread で保持される）。
3. 描画: [AskAiPage](../frontend/src/pages/AskAiPage.tsx) の `messages.map` が **`key={message.clientId ?? message.id}`** で
   [MessageBubble](../frontend/src/components/MessageBubble.tsx) を並べる。key を clientId で安定させるのは、
   done の id 差し替えでバブルが remount して**ペーシングの残り放出が全文ジャンプになる**のを防ぐため（FRESTYLE-146）。

## Gemini 実物の仕様（一次調査の結果）

FRESTYLE-146 で Gemini web 本体の配信バンドル（gemini.gstatic.com）原文と、headless ブラウザで実際に応答を
ストリーミングさせた計測（`document.getAnimations()` サンプリング）から、実物のアルゴリズムを特定した:

- **放出単位**: 本文を句読点 regex `/[,:\.!\?、。，۔]+/`（sub-sentence 粒度 = 読点でも切る）で分割し、
  `<span class="pending">`（display:none）として先行挿入。
- **放出機構**: rAF ループが適応間隔 **fb = (チャンク到着間隔の逐次平均 EMA + 600) / (残チャンク数 + 1)**
  （EMA サンプルは 1000ms キャップ・初期 3000ms）で 1 チャンクずつ `.pending`→`.animating` に付け替え。
  遅延が fb×10 を超えたら floor(遅延/fb) 個を一括表示（catch-up）。応答完了（COMPLETE）で EMA×0.8 に加速。
- **アニメ**: `@keyframes fade-in-text { 0%{opacity:0} to{opacity:1} }` — **opacity のみ（blur / translate 無し）**。
  実測 400ms / ease-out / fill forwards / 1 回。

## FreStyle での再現実装

### ペーシング: [useSmoothReveal](../frontend/src/hooks/useSmoothReveal.ts)

MessageBubble 内で `content`（真の受信全文）から「表示してよい prefix」を Gemini と同じ適応リズムで進める。
messages state には手を入れず、**表示ポリシーだけを component 層で持つ**（done の id 差し替え・error 置換・
streaming 中 fetchMessages skip の race guard をすべて無傷に保つため）。

- 放出単位は [subSentenceSegments](../frontend/src/components/message/subSentenceSegments.ts) の句読点チャンク
  （Gemini の regex ＋ **改行も境界に追加** — コードブロック等の句読点が無い行でも行単位で放出できるようにする独自拡張）。
- 放出間隔・catch-up・完了時の加速は Gemini の式をそのまま実装（fb = (EMA+600)/(pending+1)、fb×10 超で一括、×0.8）。
- **mount 時は全文即時表示**（履歴ロード・確定メッセージ・既存テストをこのルール 1 つで無変更に保つ）。
- **未完チャンクの保留**: 句読点で終わっていない末尾は streaming 中は出さない（出すと次の delta で span テキストが
  書き換わり、フェード無しで文字が増える）。120 文字を超えたら停滞防止のため放出。完了時は全部出し切る。
- `content` が表示済み prefix の延長でないとき（既に一部表示済みで SSE error に全文置換された等）は**即スナップ**。
- 完了（`active` true→false）後は残りをリズム付きで流し切り、`settled` になってから bookend favicon を出す。

**`done` 以外の終端（error / 連投 abort）の扱い**（[useAskAi](../frontend/src/hooks/useAskAi.ts)）: これらは placeholder の
`id` を `streaming-` のままにすると `active` が永久に true に張り付き、句読点を含まないエラー文が未完チャンク保留で
永久に非表示になる。そのため error ハンドラは `id` を `error-…` に、連投時の旧 placeholder は `aborted-…` に差し替えて
**ストリーミング状態を閉じる**（`clientId` は保持されるので key は安定し remount しない）。id が非 streaming になると
`active` が false になり、useSmoothReveal の drain 経路が受信済みぶん（未完チャンク含む）を流し切る。

### フェード: [rehypeFadeSegments](../frontend/src/components/message/rehypeFadeSegments.ts) + [index.css](../frontend/src/index.css)

- ストリーミング中だけ、react-markdown の rehype 段でテキストノードを**句読点チャンクの `<span class="fade-seg">`**
  に分割（`pre`/`code` 配下は分割しない）。放出境界とチャンク境界が一致するので、span テキストの書き換えは起きない。
- `.fade-seg { animation: fade-seg-in 400ms ease-out both }` — **Gemini 実測値そのまま（opacity のみ）**。
  blur・translate は実物に存在しないため使わない（FRESTYLE-145 の blur は 146 で撤去）。
- 既表示チャンクは React が DOM ノードを再利用するためアニメが再生し直されない。新しく mount した span だけが
  1 回フェードする（[MarkdownView.test](../frontend/src/components/message/__tests__/MarkdownView.test.tsx) が
  DOM ノード同一性でガード）。
- `prefers-reduced-motion: reduce` は `animation-duration: 0.01ms`（実質即時。`animationend` を残すため 0 にしない）。
- 完了後（settled）はプラグインを外し素の Markdown（span ゼロ）に戻す。[MarkdownView](../frontend/src/components/message/MarkdownView.tsx)
  は memo 化してあり、再パースは「SSE token 毎」ではなく「放出毎」に抑えられる。

### 自動スクロール

ペーシングの放出は messages state と非同期に DOM の高さを伸ばすため、既存の `[messages]` 依存 effect に加えて
**ResizeObserver**（jsdom には無いので存在ガード付き）でリストの高さ変化を拾い、`stickToBottomRef` が立っている
ときだけ最下部へ追従する。ユーザーが上へスクロールしたら追従しない挙動（wheel/touchmove 検知）は従来のまま。

## 調整パラメータ（体感チューニングの入口）

| 場所 | 値 | 意味 |
|---|---|---|
| `useSmoothReveal.ts` `FB_BASE_MS` | 600 | fb 式の基底。大きいほどゆっくり |
| `useSmoothReveal.ts` `EMA_INITIAL_MS` / `EMA_SAMPLE_CAP_MS` | 3000 / 1000 | 到着間隔平均の初期値・上限 |
| `useSmoothReveal.ts` `COMPLETE_ACCEL` | 0.8 | 完了後の加速率 |
| `useSmoothReveal.ts` `MAX_TAIL_HOLD_CHARS` | 120 | 未完チャンク保留の上限 |
| `index.css` `.fade-seg` | 400ms ease-out | フェードの長さ（Gemini 実測値） |

## テスト

- [useSmoothReveal.test](../frontend/src/hooks/__tests__/useSmoothReveal.test.ts): fake timers で mount 即時全文 /
  初回チャンク即時 / タイマー放出 / 未完チャンク保留 / fast-drain / エラー時スナップ / catch-up / unmount cleanup。
- [rehypeFadeSegments.test](../frontend/src/components/message/__tests__/rehypeFadeSegments.test.ts):
  句読点チャンク分割・全文保存・`pre`/`code` 非分割・空白の扱い。
- [MessageBubble.test](../frontend/src/components/__tests__/MessageBubble.test.tsx): streaming id での span 付与・
  ペーシングの時間経過検証・bookend の出し分け。
- [useAskAi.test](../frontend/src/hooks/__tests__/useAskAi.test.tsx): done 後の clientId 保持（key 安定化の前提）。
