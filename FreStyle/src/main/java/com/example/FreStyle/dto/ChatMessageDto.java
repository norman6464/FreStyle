// package com.example.FreStyle.dto;

// import lombok.AllArgsConstructor;
// import lombok.Data;
// import lombok.NoArgsConstructor;

// @Data
// @NoArgsConstructor
// @AllArgsConstructor
// public class ChatMessageDto {
//   private String content;
//   private boolean isUser;
//   private long timestamp;
// }


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
    private Integer senderId;
    private String senderName;
    private String content;   // ここにメッセージ本文を入れる場合
    private Timestamp createdAt;
    private Timestamp updatedAt;
}
