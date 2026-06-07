package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** AI チャットセッションのタイトル更新の入力。 */
public record UpdateSessionTitleRequest(@NotBlank String title) {}
