package com.example.FreStyle.dto;

public record DailyGoalDto(
        String date,
        Integer target,
        Integer completed) {
}
