package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** コード実行(サンドボックス)のリクエスト。language 省略時は java とみなす。 */
public record CodeExecuteRequest(@NotBlank String code, String language) {}
