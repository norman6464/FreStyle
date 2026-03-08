package com.example.FreStyle.form;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Pattern;

public record ReminderSettingForm(
    @NotNull Boolean enabled,
    @NotBlank @Pattern(regexp = "^([01]\\d|2[0-3]):[0-5]\\d$") String reminderTime,
    @NotBlank String daysOfWeek
) {}
