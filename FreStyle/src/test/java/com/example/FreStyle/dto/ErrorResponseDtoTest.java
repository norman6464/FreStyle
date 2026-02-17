package com.example.FreStyle.dto;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.time.LocalDateTime;

import static org.assertj.core.api.Assertions.assertThat;

class ErrorResponseDtoTest {

    @Test
    @DisplayName("ファクトリーメソッドofでタイムスタンプが自動設定される")
    void ofSetsTimestamp() {
        LocalDateTime before = LocalDateTime.now();

        ErrorResponseDto dto = ErrorResponseDto.of(404, "Not Found", "リソースが見つかりません", "/api/test");

        assertThat(dto.timestamp()).isNotNull();
        assertThat(dto.timestamp()).isAfterOrEqualTo(before);
        assertThat(dto.timestamp()).isBeforeOrEqualTo(LocalDateTime.now());
    }

    @Test
    @DisplayName("ファクトリーメソッドofで全フィールドが正しく設定される")
    void ofSetsAllFields() {
        ErrorResponseDto dto = ErrorResponseDto.of(400, "Bad Request", "不正なリクエストです", "/api/auth/login");

        assertThat(dto.status()).isEqualTo(400);
        assertThat(dto.error()).isEqualTo("Bad Request");
        assertThat(dto.message()).isEqualTo("不正なリクエストです");
        assertThat(dto.path()).isEqualTo("/api/auth/login");
    }

    @Test
    @DisplayName("recordコンストラクタで全フィールドが設定される")
    void recordConstructorWorks() {
        LocalDateTime now = LocalDateTime.now();
        ErrorResponseDto dto = new ErrorResponseDto(now, 500, "Internal Server Error", "サーバーエラー", "/api/test");

        assertThat(dto.timestamp()).isEqualTo(now);
        assertThat(dto.status()).isEqualTo(500);
        assertThat(dto.error()).isEqualTo("Internal Server Error");
        assertThat(dto.message()).isEqualTo("サーバーエラー");
        assertThat(dto.path()).isEqualTo("/api/test");
    }
}
