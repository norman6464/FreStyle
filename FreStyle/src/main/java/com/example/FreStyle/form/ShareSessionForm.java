package com.example.FreStyle.form;

import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;

public record ShareSessionForm(
    @NotNull Integer sessionId,
    @Size(max = 1000) String description
) {}
