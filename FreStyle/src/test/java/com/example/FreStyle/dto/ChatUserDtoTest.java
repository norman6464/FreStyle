package com.example.FreStyle.dto;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.sql.Timestamp;

import static org.assertj.core.api.Assertions.assertThat;

class ChatUserDtoTest {

    @Test
    @DisplayName("簡易コンストラクタでunreadCountが0に初期化される")
    void simpleConstructorInitializesUnreadCountToZero() {
        ChatUserDto dto = new ChatUserDto(1, "test@example.com", "テスト", 10);

        assertThat(dto.getUserId()).isEqualTo(1);
        assertThat(dto.getEmail()).isEqualTo("test@example.com");
        assertThat(dto.getName()).isEqualTo("テスト");
        assertThat(dto.getRoomId()).isEqualTo(10);
        assertThat(dto.getUnreadCount()).isZero();
    }

    @Test
    @DisplayName("簡易コンストラクタでメッセージ関連フィールドがnull")
    void simpleConstructorLeavesMessageFieldsNull() {
        ChatUserDto dto = new ChatUserDto(1, "test@example.com", "テスト", 10);

        assertThat(dto.getLastMessage()).isNull();
        assertThat(dto.getLastMessageSenderId()).isNull();
        assertThat(dto.getLastMessageSenderName()).isNull();
        assertThat(dto.getLastMessageAt()).isNull();
        assertThat(dto.getProfileImage()).isNull();
    }

    @Test
    @DisplayName("AllArgsConstructorで全フィールドが設定される")
    void allArgsConstructorSetsAllFields() {
        Timestamp now = new Timestamp(System.currentTimeMillis());
        ChatUserDto dto = new ChatUserDto(
                1, "test@example.com", "テスト", 10,
                "最後のメッセージ", 2, "送信者", now, 3, "https://example.com/image.jpg");

        assertThat(dto.getUserId()).isEqualTo(1);
        assertThat(dto.getLastMessage()).isEqualTo("最後のメッセージ");
        assertThat(dto.getLastMessageSenderId()).isEqualTo(2);
        assertThat(dto.getLastMessageSenderName()).isEqualTo("送信者");
        assertThat(dto.getLastMessageAt()).isEqualTo(now);
        assertThat(dto.getUnreadCount()).isEqualTo(3);
        assertThat(dto.getProfileImage()).isEqualTo("https://example.com/image.jpg");
    }
}
