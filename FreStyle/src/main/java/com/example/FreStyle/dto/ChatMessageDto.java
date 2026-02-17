// NoSQL → RDBを使用するので変更した
package com.example.FreStyle.dto;

import java.sql.Timestamp;

public record ChatMessageDto(
        Integer id,
        Integer roomId,
        Integer senderId,
        String senderName,
        String content,
        Timestamp createdAt,
        Timestamp updatedAt
) {}
