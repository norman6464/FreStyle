package com.example.FreStyle.dto;

public record WeeklyChallengeDto(
    Integer id,
    String title,
    String description,
    String category,
    int targetSessions,
    int completedSessions,
    boolean isCompleted,
    String weekStart,
    String weekEnd
) {}
