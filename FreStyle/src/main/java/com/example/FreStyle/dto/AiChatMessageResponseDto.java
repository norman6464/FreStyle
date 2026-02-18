package com.example.FreStyle.dto;

public record AiChatMessageResponseDto(
        String id,
        Integer sessionId,
        Integer userId,
        String role,
        String content,
        Long createdAt
) {}
