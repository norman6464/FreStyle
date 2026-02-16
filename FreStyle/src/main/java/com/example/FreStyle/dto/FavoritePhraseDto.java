package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class FavoritePhraseDto {
    private Integer id;
    private String originalText;
    private String rephrasedText;
    private String pattern;
    private String createdAt;
}
