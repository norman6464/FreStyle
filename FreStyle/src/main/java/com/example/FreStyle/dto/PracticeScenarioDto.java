package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class PracticeScenarioDto {
    private Integer id;
    private String name;
    private String description;
    private String category;
    private String roleName;
    private String difficulty;
    private String systemPrompt;
}
