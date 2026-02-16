package com.example.FreStyle.dto;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class UserDtoTest {

    @Test
    @DisplayName("3引数コンストラクタでroomIdがnull")
    void threeArgConstructorLeavesRoomIdNull() {
        UserDto dto = new UserDto(1, "test@example.com", "テストユーザー");

        assertThat(dto.getId()).isEqualTo(1);
        assertThat(dto.getEmail()).isEqualTo("test@example.com");
        assertThat(dto.getName()).isEqualTo("テストユーザー");
        assertThat(dto.getRoomId()).isNull();
    }

    @Test
    @DisplayName("AllArgsConstructorでroomIdが設定される")
    void allArgsConstructorSetsRoomId() {
        UserDto dto = new UserDto(1, "test@example.com", "テストユーザー", 5);

        assertThat(dto.getId()).isEqualTo(1);
        assertThat(dto.getRoomId()).isEqualTo(5);
    }

    @Test
    @DisplayName("setterでroomIdを後から設定できる")
    void setRoomIdAfterConstruction() {
        UserDto dto = new UserDto(1, "test@example.com", "テストユーザー");
        assertThat(dto.getRoomId()).isNull();

        dto.setRoomId(10);
        assertThat(dto.getRoomId()).isEqualTo(10);
    }

    @Test
    @DisplayName("NoArgsConstructorで全フィールドがnull")
    void noArgsConstructorAllFieldsNull() {
        UserDto dto = new UserDto();

        assertThat(dto.getId()).isNull();
        assertThat(dto.getEmail()).isNull();
        assertThat(dto.getName()).isNull();
        assertThat(dto.getRoomId()).isNull();
    }

    @Test
    @DisplayName("setterでemail・nameを変更できる")
    void setEmailAndName() {
        UserDto dto = new UserDto(1, "old@example.com", "旧名前");

        dto.setEmail("new@example.com");
        dto.setName("新名前");

        assertThat(dto.getEmail()).isEqualTo("new@example.com");
        assertThat(dto.getName()).isEqualTo("新名前");
    }
}
