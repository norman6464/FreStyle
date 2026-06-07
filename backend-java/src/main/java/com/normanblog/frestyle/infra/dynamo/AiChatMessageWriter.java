package com.normanblog.frestyle.infra.dynamo;

import com.normanblog.frestyle.dto.AiChatMessageResponse;

/** AI チャットメッセージ(DynamoDB)の書き込み境界。 */
public interface AiChatMessageWriter {

  /** メッセージ 1 件を保存する(PutItem)。 */
  void save(AiChatMessageResponse message);
}
