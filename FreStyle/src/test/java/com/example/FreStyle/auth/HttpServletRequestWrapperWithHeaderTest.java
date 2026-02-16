package com.example.FreStyle.auth;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.Collections;
import java.util.Enumeration;
import java.util.List;

import org.junit.jupiter.api.Test;

import jakarta.servlet.http.HttpServletRequest;

class HttpServletRequestWrapperWithHeaderTest {

    @Test
    void カスタムヘッダーが取得できる() {
        HttpServletRequest original = mock(HttpServletRequest.class);
        HttpServletRequestWrapperWithHeader wrapper =
                new HttpServletRequestWrapperWithHeader(original, "Authorization", "Bearer token123");

        assertEquals("Bearer token123", wrapper.getHeader("Authorization"));
    }

    @Test
    void 元のヘッダーがカスタムヘッダーに含まれない場合はオリジナルから取得する() {
        HttpServletRequest original = mock(HttpServletRequest.class);
        when(original.getHeader("Content-Type")).thenReturn("application/json");

        HttpServletRequestWrapperWithHeader wrapper =
                new HttpServletRequestWrapperWithHeader(original, "Authorization", "Bearer token");

        assertEquals("application/json", wrapper.getHeader("Content-Type"));
    }

    @Test
    void getHeaderNamesにカスタムヘッダーが含まれる() {
        HttpServletRequest original = mock(HttpServletRequest.class);
        when(original.getHeaderNames()).thenReturn(Collections.enumeration(List.of("Content-Type")));

        HttpServletRequestWrapperWithHeader wrapper =
                new HttpServletRequestWrapperWithHeader(original, "Authorization", "Bearer token");

        Enumeration<String> names = wrapper.getHeaderNames();
        List<String> nameList = Collections.list(names);

        assertTrue(nameList.contains("Authorization"));
        assertTrue(nameList.contains("Content-Type"));
    }
}
