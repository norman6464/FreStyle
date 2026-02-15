package com.example.FreStyle.dto;

public record PresignedUrlResponse(
    String uploadUrl,
    String imageUrl
) {}
