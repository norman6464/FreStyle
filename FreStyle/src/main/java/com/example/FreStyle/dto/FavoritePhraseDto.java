package com.example.FreStyle.dto;

public record FavoritePhraseDto(
        Integer id,
        String originalText,
        String rephrasedText,
        String pattern,
        String createdAt) {
}
