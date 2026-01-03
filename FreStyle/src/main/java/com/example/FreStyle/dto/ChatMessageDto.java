// NoSQL → RDBを使用するので変更した
package com.example.FreStyle.dto;

import java.sql.Timestamp;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ChatMessageDto {
    private Integer id;
    private Integer roomId;
    private Integer senderId;  // 送信者のユーザーID（内部用）
    private String senderName;
    private String content;
    private Timestamp createdAt;
    private Timestamp updatedAt;
}
