package com.example.FreStyle.dto;

public record PracticeScenarioDto(
        Integer id,
        String name,
        String description,
        String category,
        String roleName,
        String difficulty,
        String systemPrompt
) {}
