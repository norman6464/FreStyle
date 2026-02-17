package com.example.FreStyle.dto;

import java.time.LocalDateTime;

public record ErrorResponseDto(
        LocalDateTime timestamp,
        int status,
        String error,
        String message,
        String path) {

    public static ErrorResponseDto of(int status, String error, String message, String path) {
        return new ErrorResponseDto(
            LocalDateTime.now(),
            status,
            error,
            message,
            path
        );
    }
}
