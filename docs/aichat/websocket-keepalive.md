# AI チャット WebSocket — keepalive と自動再接続

## 背景

ユーザーが AI チャット画面を開いたまましばらく放置 → 次のメッセージを送っても返信が来ない、という不具合があった。

原因:

1. **ALB のアイドルタイムアウト**（既定 60 秒）で WebSocket が切断される
2. バックエンドに ping/pong がなく、放置中の接続を ALB が「死んだ」と判断する
3. フロントは onclose で何もしないため、切断後にユーザーが新しいメッセージを送っても、すでに closed の socket にキューが積まれるだけで送信されない
4. バックエンドの usecase 呼び出しが `c.Request.Context()` を使っており、接続切断時に正しく cancel できない可能性
5. Bedrock client の初期化失敗時 `sendMsg` が nil で渡され、メッセージ送信のたびに 500 / panic

## 修正概要

### Backend (`backend/internal/handler/ai_chat_ws_handler.go`)

- **ping/pong keepalive**: 54 秒ごとに server → client へ ping を送り、ALB のアイドルタイムアウトを回避
- **pong wait**: 60 秒以内に pong（または通常メッセージ）が来なければ切断
- **書き込み単一化**: gorilla/websocket は単一接続を複数 goroutine から書けないため、書き込み専用 goroutine（`writePump`）に集約
- **独立 context**: `context.WithCancel(context.Background())` を作って ws lifecycle に紐付け、`c.Request.Context()` の cancel に巻き込まれない
- **Bedrock 不在時の 503**: `sendMsg == nil` のとき upgrade 前に 503 を返してサイレント切断を回避

```
const (
    aiChatWriteWait      = 10 * time.Second
    aiChatPongWait       = 60 * time.Second
    aiChatPingPeriod     = (aiChatPongWait * 9) / 10  // 54s
    aiChatMaxMessageSize = 64 * 1024
)
```

### Frontend (`frontend/src/hooks/useWebSocketNative.ts`)

- **自動再接続**: onclose で指数バックオフ（1s → 2s → 4s → 8s → 16s → 30s 上限）で再接続
- **キュー保持**: 切断中の `send()` は `queueRef` に積まれ、再接続成功後の onopen で `flushQueue` により送信
- **disconnect 経由は再接続しない**: ユーザーが明示的にログアウト等で切った場合は `cancelledRef = true` でループを止める

```
sock.onclose = () => {
  onCloseRef.current?.();
  if (cancelledRef.current) return;
  const delay = Math.min(30000, 1000 * 2 ** attempt);
  reconnectTimerRef.current = setTimeout(connect, delay);
};
```

## テスト

| ケース | 場所 |
|---|---|
| ALB が切ったケースで onclose 後に再接続が走る | `useWebSocketNative.test.ts` |
| `disconnect()` 後は再接続しない | 同上 |
| OPEN 前 send → キュー → open 後 flush | 同上（既存） |

バックエンドの ping ループは時間依存テストが書きづらいため、本 PR では単体テストではなく実機検証で確認する想定。

## 動作確認手順

1. `docker compose up -d --build` でローカル起動
2. ブラウザで AI チャット画面を開いて 1 通送信 → 返信を確認
3. **画面を 2〜3 分放置** → 4 通目を送信 → ALB タイムアウトを再現できる本番環境でも返信が来ること
4. ブラウザの DevTools → Network → WS を開き、Frames タブで ping (opcode 0x9) / pong (0xA) が周期的に流れているか確認
