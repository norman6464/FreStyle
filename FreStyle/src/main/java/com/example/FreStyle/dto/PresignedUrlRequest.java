package com.example.FreStyle.dto;

import jakarta.validation.constraints.NotBlank;

public record PresignedUrlRequest(
    @NotBlank String fileName,
    @NotBlank String contentType
) {}
