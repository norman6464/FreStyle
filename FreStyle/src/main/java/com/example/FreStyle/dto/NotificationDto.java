package com.example.FreStyle.dto;

public record NotificationDto(
        Integer id,
        String type,
        String title,
        String message,
        Boolean isRead,
        Integer relatedId,
        String createdAt) {
}
