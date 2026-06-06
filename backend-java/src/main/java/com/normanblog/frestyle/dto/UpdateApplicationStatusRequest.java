package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** 企業利用申請の status 更新リクエスト(super_admin)。 */
public record UpdateApplicationStatusRequest(@NotBlank String status) {}
