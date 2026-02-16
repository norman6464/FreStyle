package com.example.FreStyle.exception;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

@DisplayName("BusinessException テスト")
class BusinessExceptionTest {

    @Test
    @DisplayName("メッセージ付きコンストラクタでメッセージを保持する")
    void constructor_WithMessage() {
        BusinessException exception = new BusinessException("テストエラー");

        assertEquals("テストエラー", exception.getMessage());
        assertNull(exception.getCause());
    }

    @Test
    @DisplayName("メッセージ・原因例外付きコンストラクタで両方を保持する")
    void constructor_WithMessageAndCause() {
        RuntimeException cause = new RuntimeException("原因");
        BusinessException exception = new BusinessException("テストエラー", cause);

        assertEquals("テストエラー", exception.getMessage());
        assertEquals(cause, exception.getCause());
    }

    @Test
    @DisplayName("RuntimeExceptionのサブクラスである")
    void isRuntimeException() {
        BusinessException exception = new BusinessException("テスト");

        assertInstanceOf(RuntimeException.class, exception);
    }
}
