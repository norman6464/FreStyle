package com.example.FreStyle.form;

import jakarta.validation.constraints.NotNull;

public record ShareSessionForm(
    @NotNull Integer sessionId,
    String description
) {}
