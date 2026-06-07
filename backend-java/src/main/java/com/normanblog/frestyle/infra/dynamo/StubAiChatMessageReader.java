package com.normanblog.frestyle.infra.dynamo;

import com.normanblog.frestyle.dto.AiChatMessageResponse;
import java.util.List;

/**
 * DynamoDB を呼ばない stub。table 未設定のローカル / テスト環境で使う。
 *
 * <p>常に空リストを返す(メッセージ履歴なし扱い)。本番では {@link DynamoAiChatMessageReader} に
 * 差し替わる。
 */
public class StubAiChatMessageReader implements AiChatMessageReader {

  @Override
  public List<AiChatMessageResponse> listBySession(Long sessionId) {
    return List.of();
  }
}
