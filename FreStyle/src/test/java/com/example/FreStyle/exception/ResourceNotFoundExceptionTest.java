package com.example.FreStyle.exception;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

@DisplayName("ResourceNotFoundException テスト")
class ResourceNotFoundExceptionTest {

    @Test
    @DisplayName("メッセージ付きコンストラクタでメッセージを保持する")
    void constructor_WithMessage() {
        ResourceNotFoundException exception = new ResourceNotFoundException("リソース未検出");

        assertEquals("リソース未検出", exception.getMessage());
        assertNull(exception.getCause());
    }

    @Test
    @DisplayName("メッセージ・原因例外付きコンストラクタで両方を保持する")
    void constructor_WithMessageAndCause() {
        RuntimeException cause = new RuntimeException("原因");
        ResourceNotFoundException exception = new ResourceNotFoundException("リソース未検出", cause);

        assertEquals("リソース未検出", exception.getMessage());
        assertEquals(cause, exception.getCause());
    }

    @Test
    @DisplayName("RuntimeExceptionのサブクラスである")
    void isRuntimeException() {
        ResourceNotFoundException exception = new ResourceNotFoundException("テスト");

        assertInstanceOf(RuntimeException.class, exception);
    }

    @Test
    @DisplayName("nullメッセージでもnullが保持される")
    void constructor_WithNullMessage() {
        ResourceNotFoundException exception = new ResourceNotFoundException(null);

        assertNull(exception.getMessage());
    }

    @Test
    @DisplayName("2引数コンストラクタでcauseがnullでも正常に動作する")
    void constructor_WithNullCause() {
        ResourceNotFoundException exception = new ResourceNotFoundException("エラー", null);

        assertEquals("エラー", exception.getMessage());
        assertNull(exception.getCause());
    }
}
