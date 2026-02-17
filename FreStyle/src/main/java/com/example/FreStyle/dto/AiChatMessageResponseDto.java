package com.example.FreStyle.dto;

import java.sql.Timestamp;

public record AiChatMessageResponseDto(
        Integer id,
        Integer sessionId,
        Integer userId,
        String role,
        String content,
        Timestamp createdAt
) {}
