package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** POST /exercises/{slug}/submit の入力。 */
public record SubmitExerciseRequest(@NotBlank String code) {}
