package com.normanblog.frestyle.dynamo;

import com.normanblog.frestyle.dto.AiChatAttachmentDto;
import com.normanblog.frestyle.dto.AiChatMessageResponse;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;
import software.amazon.awssdk.services.dynamodb.model.QueryRequest;
import software.amazon.awssdk.services.dynamodb.model.QueryResponse;

/**
 * DynamoDB から AI チャットメッセージを読む本番実装。
 *
 * <p>テーブルスキーマは Go 版と同一: PK {@code sessionId}(数値を文字列化) / SK {@code messageId} /
 * 属性 role・content・createdAt(RFC3339)・attachments(マップのリスト)。messageId 昇順で返す。
 */
public class DynamoAiChatMessageReader implements AiChatMessageReader, AutoCloseable {

  private final DynamoDbClient client;
  private final String table;

  public DynamoAiChatMessageReader(DynamoDbClient client, String table) {
    this.client = client;
    this.table = table;
  }

  // DynamoDbClient は Closeable。Spring の destroy-method 推論でコンテキスト終了時に閉じる。
  @Override
  public void close() {
    client.close();
  }

  @Override
  public List<AiChatMessageResponse> listBySession(Long sessionId) {
    QueryRequest request =
        QueryRequest.builder()
            .tableName(table)
            .keyConditionExpression("sessionId = :sid")
            .expressionAttributeValues(
                Map.of(":sid", AttributeValue.fromS(String.valueOf(sessionId))))
            // SK(messageId)昇順 = 作成順。
            .scanIndexForward(true)
            .build();
    QueryResponse response = client.query(request);

    List<AiChatMessageResponse> messages = new ArrayList<>(response.items().size());
    for (Map<String, AttributeValue> item : response.items()) {
      messages.add(toMessage(sessionId, item));
    }

    return messages;
  }

  private AiChatMessageResponse toMessage(Long sessionId, Map<String, AttributeValue> item) {
    return new AiChatMessageResponse(
        sessionId,
        str(item.get("messageId")),
        str(item.get("role")),
        str(item.get("content")),
        toAttachments(item.get("attachments")),
        str(item.get("createdAt")));
  }

  private List<AiChatAttachmentDto> toAttachments(AttributeValue value) {
    if (value == null || !value.hasL() || value.l().isEmpty()) {
      return null;
    }
    List<AiChatAttachmentDto> attachments = new ArrayList<>(value.l().size());
    for (AttributeValue element : value.l()) {
      Map<String, AttributeValue> m = element.m();
      attachments.add(
          new AiChatAttachmentDto(
              str(m.get("key")),
              str(m.get("filename")),
              str(m.get("contentType")),
              str(m.get("format")),
              str(m.get("kind")),
              num(m.get("sizeBytes"))));
    }

    return attachments;
  }

  private static String str(AttributeValue value) {
    return value == null ? null : value.s();
  }

  private static long num(AttributeValue value) {
    if (value == null || value.n() == null) {
      return 0L;
    }
    return Long.parseLong(value.n());
  }
}
