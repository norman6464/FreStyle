package com.example.FreStyle.controller;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

@DisplayName("HelloController テスト")
class HelloControllerTest {

    private final HelloController helloController = new HelloController();

    @Test
    @DisplayName("helloエンドポイントが'hello'を返す")
    void hello_ReturnsHelloString() {
        String result = helloController.hello();

        assertEquals("hello", result);
    }
}
