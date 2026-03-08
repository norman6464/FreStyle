package com.example.FreStyle.dto;

public record SharedSessionDto(
    Integer id,
    Integer sessionId,
    String sessionTitle,
    Integer userId,
    String username,
    String userIconUrl,
    String description,
    String createdAt
) {}
