package com.example.FreStyle.dto;

public record ChatMessageDto(
        String id,
        Integer roomId,
        Integer senderId,
        String senderName,
        String content,
        Long createdAt
) {}
