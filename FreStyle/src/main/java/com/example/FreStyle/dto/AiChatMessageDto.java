package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

// RestControllerでDTOのようなオブジェクト側をJSONで返却をする場合は
// フィールド名がJSONでのプロパティ名になる
@Data
@NoArgsConstructor
@AllArgsConstructor
public class AiChatMessageDto {
  private String content;
  private boolean isUser;
  private long timestamp;

}
