package com.example.FreStyle.dto;

public record ReminderSettingDto(
    boolean enabled,
    String reminderTime,
    String daysOfWeek
) {}
