package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** AI チャットセッション作成の入力。sessionType 省略時は free 扱い。 */
public record CreateAiChatSessionRequest(
    @NotBlank String title, String sessionType, Long scenarioId) {}
