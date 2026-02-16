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
}
