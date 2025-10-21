package com.example.FreStyle.dto;


// RestControllerでDTOのようなオブジェクト側をJSONで返却をする場合は
// フィールド名がJSONでのプロパティ名になる
public class AiChatMessageDto {
  private String content;
  private boolean isUser;
  private long timestamp;

  public AiChatMessageDto() {
  }

  public AiChatMessageDto(String content, boolean isUser, long timestamp) {
    this.content = content;
    this.isUser = isUser;
    this.timestamp = timestamp;
  }

  public String getContent() {
    return content;
  }

  public boolean isUser() {
    return isUser;
  }

  public long getTimestamp() {
    return timestamp;
  }

  public void setContent(String content) {
    this.content = content;
  }

  public void setUser(boolean isUser) {
    this.isUser = isUser;
  }

  public void setTimestamp(long timestamp) {
    this.timestamp = timestamp;
  }

}
