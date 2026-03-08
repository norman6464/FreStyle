package com.example.FreStyle.dto;

public record ConversationTemplateDto(
    Integer id,
    String title,
    String description,
    String category,
    String openingMessage,
    String difficulty
) {}
