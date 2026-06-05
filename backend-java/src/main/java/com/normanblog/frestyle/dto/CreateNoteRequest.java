package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** ノート作成リクエストのボディ。 */
public record CreateNoteRequest(@NotBlank String title, String content) {}
