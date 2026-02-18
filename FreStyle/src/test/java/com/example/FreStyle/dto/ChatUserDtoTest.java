package com.example.FreStyle.dto;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

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
    @DisplayName("NoArgsConstructorで全フィールドがnull/デフォルト値")
    void noArgsConstructorLeavesAllFieldsNull() {
        ChatUserDto dto = new ChatUserDto();

        assertThat(dto.getUserId()).isNull();
        assertThat(dto.getEmail()).isNull();
        assertThat(dto.getName()).isNull();
        assertThat(dto.getRoomId()).isNull();
        assertThat(dto.getLastMessage()).isNull();
        assertThat(dto.getLastMessageSenderId()).isNull();
        assertThat(dto.getLastMessageSenderName()).isNull();
        assertThat(dto.getLastMessageAt()).isNull();
        assertThat(dto.getUnreadCount()).isNull();
        assertThat(dto.getProfileImage()).isNull();
    }

    @Test
    @DisplayName("setterで個別フィールドを更新できる")
    void setterUpdatesFields() {
        ChatUserDto dto = new ChatUserDto(1, "old@example.com", "旧名前", 10);

        dto.setName("新名前");
        dto.setProfileImage("https://example.com/new.jpg");
        dto.setUnreadCount(5);

        assertThat(dto.getName()).isEqualTo("新名前");
        assertThat(dto.getProfileImage()).isEqualTo("https://example.com/new.jpg");
        assertThat(dto.getUnreadCount()).isEqualTo(5);
    }

    @Test
    @DisplayName("AllArgsConstructorで全フィールドが設定される")
    void allArgsConstructorSetsAllFields() {
        Long now = System.currentTimeMillis();
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
