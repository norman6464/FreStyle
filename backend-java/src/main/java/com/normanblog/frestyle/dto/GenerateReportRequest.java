package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;

/** 月次レポート生成要求の入力。year + month から月初〜翌月初の期間を組み立てる。 */
public record GenerateReportRequest(
    @NotNull @Min(2000) @Max(2100) Integer year, @NotNull @Min(1) @Max(12) Integer month) {}
