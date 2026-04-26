package com.example.FreStyle.form;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

/**
 * 管理者によるシナリオ作成・更新時の入力フォーム。
 */
public record PracticeScenarioForm(
        @NotBlank @Size(max = 100) String name,
        @Size(max = 1000) String description,
        @Size(max = 50) String category,
        @NotBlank @Size(max = 100) String roleName,
        @Size(max = 20) String difficulty,
        @NotBlank String systemPrompt
) {}
