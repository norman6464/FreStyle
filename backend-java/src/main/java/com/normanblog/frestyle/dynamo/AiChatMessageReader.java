package com.normanblog.frestyle.dynamo;

import com.normanblog.frestyle.dto.AiChatMessageResponse;
import java.util.List;

/** AI チャットメッセージ(DynamoDB)の読み取り境界。 */
public interface AiChatMessageReader {

  /** 指定セッションのメッセージを作成順(messageId 昇順)で返す。 */
  List<AiChatMessageResponse> listBySession(Long sessionId);
}
