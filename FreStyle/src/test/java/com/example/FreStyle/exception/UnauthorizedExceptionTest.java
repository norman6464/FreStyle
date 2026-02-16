package com.example.FreStyle.exception;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

@DisplayName("UnauthorizedException テスト")
class UnauthorizedExceptionTest {

    @Test
    @DisplayName("メッセージ付きコンストラクタでメッセージを保持する")
    void constructor_WithMessage() {
        UnauthorizedException exception = new UnauthorizedException("権限エラー");

        assertEquals("権限エラー", exception.getMessage());
        assertNull(exception.getCause());
    }

    @Test
    @DisplayName("メッセージ・原因例外付きコンストラクタで両方を保持する")
    void constructor_WithMessageAndCause() {
        RuntimeException cause = new RuntimeException("原因");
        UnauthorizedException exception = new UnauthorizedException("権限エラー", cause);

        assertEquals("権限エラー", exception.getMessage());
        assertEquals(cause, exception.getCause());
    }

    @Test
    @DisplayName("RuntimeExceptionのサブクラスである")
    void isRuntimeException() {
        UnauthorizedException exception = new UnauthorizedException("テスト");

        assertInstanceOf(RuntimeException.class, exception);
    }

    @Test
    @DisplayName("nullメッセージでもnullが保持される")
    void constructor_WithNullMessage() {
        UnauthorizedException exception = new UnauthorizedException(null);

        assertNull(exception.getMessage());
    }

    @Test
    @DisplayName("2引数コンストラクタでcauseがnullでも正常に動作する")
    void constructor_WithNullCause() {
        UnauthorizedException exception = new UnauthorizedException("エラー", null);

        assertEquals("エラー", exception.getMessage());
        assertNull(exception.getCause());
    }
}
