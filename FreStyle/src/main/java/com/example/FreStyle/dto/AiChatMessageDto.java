package com.example.FreStyle.dto;

// RestControllerでDTOのようなオブジェクト側をJSONで返却をする場合は
// フィールド名がJSONでのプロパティ名になる
public record AiChatMessageDto(
        String content,
        boolean isUser,
        long timestamp
) {}
