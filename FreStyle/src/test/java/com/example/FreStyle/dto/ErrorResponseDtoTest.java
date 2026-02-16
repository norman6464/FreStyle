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

        assertThat(dto.getTimestamp()).isNotNull();
        assertThat(dto.getTimestamp()).isAfterOrEqualTo(before);
        assertThat(dto.getTimestamp()).isBeforeOrEqualTo(LocalDateTime.now());
    }

    @Test
    @DisplayName("ファクトリーメソッドofで全フィールドが正しく設定される")
    void ofSetsAllFields() {
        ErrorResponseDto dto = ErrorResponseDto.of(400, "Bad Request", "不正なリクエストです", "/api/auth/login");

        assertThat(dto.getStatus()).isEqualTo(400);
        assertThat(dto.getError()).isEqualTo("Bad Request");
        assertThat(dto.getMessage()).isEqualTo("不正なリクエストです");
        assertThat(dto.getPath()).isEqualTo("/api/auth/login");
    }

    @Test
    @DisplayName("AllArgsConstructorで全フィールドが設定される")
    void allArgsConstructorWorks() {
        LocalDateTime now = LocalDateTime.now();
        ErrorResponseDto dto = new ErrorResponseDto(now, 500, "Internal Server Error", "サーバーエラー", "/api/test");

        assertThat(dto.getTimestamp()).isEqualTo(now);
        assertThat(dto.getStatus()).isEqualTo(500);
        assertThat(dto.getError()).isEqualTo("Internal Server Error");
        assertThat(dto.getMessage()).isEqualTo("サーバーエラー");
        assertThat(dto.getPath()).isEqualTo("/api/test");
    }

    @Test
    @DisplayName("NoArgsConstructorで空のDTOが作成される")
    void noArgsConstructorWorks() {
        ErrorResponseDto dto = new ErrorResponseDto();

        assertThat(dto.getTimestamp()).isNull();
        assertThat(dto.getStatus()).isZero();
        assertThat(dto.getError()).isNull();
        assertThat(dto.getMessage()).isNull();
        assertThat(dto.getPath()).isNull();
    }
}
