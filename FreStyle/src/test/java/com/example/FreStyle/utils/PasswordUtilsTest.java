package com.example.FreStyle.utils;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.Test;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

class PasswordUtilsTest {

    private final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();

    @Test
    void ハッシュが生成される() {
        String hash = PasswordUtils.hash("password123");

        assertNotNull(hash);
        assertFalse(hash.isEmpty());
    }

    @Test
    void ハッシュが元のパスワードと一致する() {
        String password = "mySecretPassword";
        String hash = PasswordUtils.hash(password);

        assertTrue(encoder.matches(password, hash));
    }

    @Test
    void 異なるパスワードのハッシュは一致しない() {
        String hash = PasswordUtils.hash("correctPassword");

        assertFalse(encoder.matches("wrongPassword", hash));
    }

    @Test
    void 同じパスワードでも毎回異なるハッシュが生成される() {
        String password = "samePassword";
        String hash1 = PasswordUtils.hash(password);
        String hash2 = PasswordUtils.hash(password);

        assertNotEquals(hash1, hash2);
        // ただし両方とも元のパスワードと一致する
        assertTrue(encoder.matches(password, hash1));
        assertTrue(encoder.matches(password, hash2));
    }
}
