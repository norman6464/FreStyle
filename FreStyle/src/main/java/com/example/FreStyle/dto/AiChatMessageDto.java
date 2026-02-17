package com.example.FreStyle.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

// RestControllerでDTOのようなオブジェクト側をJSONで返却をする場合は
// フィールド名がJSONでのプロパティ名になる
public record AiChatMessageDto(
        String content,
        @JsonProperty("user") boolean isUser,
        long timestamp
) {}
