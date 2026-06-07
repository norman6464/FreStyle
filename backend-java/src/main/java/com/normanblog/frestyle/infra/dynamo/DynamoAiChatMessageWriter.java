package com.normanblog.frestyle.infra.dynamo;

import com.normanblog.frestyle.dto.AiChatAttachmentDto;
import com.normanblog.frestyle.dto.AiChatMessageResponse;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;
import software.amazon.awssdk.services.dynamodb.model.PutItemRequest;

/**
 * DynamoDB に AI チャットメッセージを書く本番実装。
 *
 * <p>テーブルスキーマは reader と同一: PK {@code sessionId}(数値を文字列化) / SK {@code messageId} /
 * 属性 role・content・createdAt(RFC3339)・attachments(マップのリスト)。
 */
public class DynamoAiChatMessageWriter implements AiChatMessageWriter, AutoCloseable {

  private final DynamoDbClient client;
  private final String table;

  public DynamoAiChatMessageWriter(DynamoDbClient client, String table) {
    this.client = client;
    this.table = table;
  }

  @Override
  public void save(AiChatMessageResponse message) {
    Map<String, AttributeValue> item = new HashMap<>();
    item.put("sessionId", AttributeValue.fromS(String.valueOf(message.sessionId())));
    item.put("messageId", AttributeValue.fromS(message.messageId()));
    item.put("role", AttributeValue.fromS(message.role()));
    item.put("content", AttributeValue.fromS(message.content() == null ? "" : message.content()));
    item.put("createdAt", AttributeValue.fromS(message.createdAt()));
    if (message.attachments() != null && !message.attachments().isEmpty()) {
      item.put("attachments", AttributeValue.fromL(toAttachmentList(message.attachments())));
    }

    client.putItem(PutItemRequest.builder().tableName(table).item(item).build());
  }

  private List<AttributeValue> toAttachmentList(List<AiChatAttachmentDto> attachments) {
    List<AttributeValue> list = new ArrayList<>(attachments.size());
    for (AiChatAttachmentDto a : attachments) {
      Map<String, AttributeValue> m = new HashMap<>();
      m.put("key", AttributeValue.fromS(nullToEmpty(a.key())));
      m.put("filename", AttributeValue.fromS(nullToEmpty(a.filename())));
      m.put("contentType", AttributeValue.fromS(nullToEmpty(a.contentType())));
      m.put("format", AttributeValue.fromS(nullToEmpty(a.format())));
      m.put("kind", AttributeValue.fromS(nullToEmpty(a.kind())));
      m.put("sizeBytes", AttributeValue.fromN(String.valueOf(a.sizeBytes())));
      list.add(AttributeValue.fromM(m));
    }

    return list;
  }

  private static String nullToEmpty(String s) {
    return s == null ? "" : s;
  }

  @Override
  public void close() {
    client.close();
  }
}
