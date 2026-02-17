package com.example.FreStyle.dto;

import java.sql.Timestamp;

public record AiChatSessionDto(
        Integer id,
        Integer userId,
        String title,
        Integer relatedRoomId,
        String scene,
        String sessionType,
        Integer scenarioId,
        Timestamp createdAt,
        Timestamp updatedAt
) {}
