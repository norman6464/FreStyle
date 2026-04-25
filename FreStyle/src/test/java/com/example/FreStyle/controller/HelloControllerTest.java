package com.example.FreStyle.controller;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.TestPropertySource;
import org.springframework.test.web.servlet.MockMvc;

/**
 * HelloController のテスト。
 *
 * <p>下記 2 段構えで Phase 1 の統合テストパターンを示す:</p>
 * <ul>
 *   <li>{@link Unit} — Controller を直接 new して戻り値を検証（軽量・高速）</li>
 *   <li>{@link Integration} — {@link MockMvc} で HTTP レイヤを通す（Security フィルタ含む）</li>
 * </ul>
 *
 * <p>Issue #1462 の Phase 1 開始テストとして、他 Controller も同パターンで追加していく。</p>
 */
class HelloControllerTest {

    @Nested
    @DisplayName("HelloController（Unit）")
    class Unit {

        private final HelloController helloController = new HelloController();

        @Test
        @DisplayName("hello() は 'hello' を返す")
        void hello_returnsHelloString() {
            String result = helloController.hello();

            assertEquals("hello", result);
        }
    }

    @Nested
    @DisplayName("HelloController（Integration: MockMvc）")
    @SpringBootTest
    @AutoConfigureMockMvc
    @TestPropertySource(locations = "classpath:application-test.properties")
    class Integration {

        @Autowired
        private MockMvc mockMvc;

        @Test
        @DisplayName("GET /api/hello は 200 + 'hello' を返す（permitAll なので認証不要）")
        void getHello_returns200() throws Exception {
            mockMvc.perform(get("/api/hello"))
                    .andExpect(status().isOk())
                    .andExpect(content().string("hello"));
        }
    }
}
