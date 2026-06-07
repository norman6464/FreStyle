package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotNull;

/** PUT /api/v2/company/settings の入力。 */
public record UpdateCompanySettingsRequest(@NotNull Boolean aiChatEnabledForTrainees) {}
