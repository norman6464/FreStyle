package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** 企業利用申請の作成リクエスト(公開フォーム)。message は任意。 */
public record CreateCompanyApplicationRequest(
    @NotBlank String companyName,
    @NotBlank String applicantName,
    @NotBlank String email,
    String message) {}
