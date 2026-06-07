package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.AiChatMessageResponse;
import com.normanblog.frestyle.entity.AiChatSession;

/**
 * AI チャット送信の進行イベントを受け取る境界。
 *
 * <p>usecase は HTTP / SSE を知らずにこの listener へ通知し、controller が SSE 書き込みに変換する
 * (レイヤード: アプリ層が HTTP の関心を持たない)。
 */
public interface AiMessageStreamListener {

  /** 新規セッションが作られたとき(最初の 1 回)。 */
  void onSession(AiChatSession session);

  /** トークン追加分が届いたとき(複数回)。 */
  void onToken(String delta);

  /** アシスタント返答が完成し保存できたとき(末尾で 1 回)。 */
  void onDone(AiChatMessageResponse finalMessage);

  /** 失敗したとき。 */
  void onError(String message);
}
