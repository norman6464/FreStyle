package com.normanblog.frestyle.dto;

/** コード実行の結果。stdout / stderr / 終了コード。 */
public record CodeExecuteResponse(String stdout, String stderr, int exitCode) {}
