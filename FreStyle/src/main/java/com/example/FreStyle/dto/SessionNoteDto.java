package com.example.FreStyle.dto;

public record SessionNoteDto(
        Integer sessionId,
        String note,
        String updatedAt) {
}
