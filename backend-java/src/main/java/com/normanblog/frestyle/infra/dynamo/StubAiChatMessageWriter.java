package com.normanblog.frestyle.infra.dynamo;

import com.normanblog.frestyle.dto.AiChatMessageResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * DynamoDB を呼ばない stub。table 未設定のローカル / テスト環境で使う。
 *
 * <p>保存をログのみに留める。本番では {@link DynamoAiChatMessageWriter} に差し替わる。
 */
public class StubAiChatMessageWriter implements AiChatMessageWriter {

  private static final Logger log = LoggerFactory.getLogger(StubAiChatMessageWriter.class);

  @Override
  public void save(AiChatMessageResponse message) {
    log.debug(
        "ai-chat message save (stub no-op): sessionId={} messageId={}",
        message.sessionId(),
        message.messageId());
  }
}
