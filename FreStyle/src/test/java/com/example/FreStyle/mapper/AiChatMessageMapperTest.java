package com.example.FreStyle.mapper;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import java.sql.Timestamp;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;

@DisplayName("AiChatMessageMapper")
class AiChatMessageMapperTest {

    private final AiChatMessageMapper mapper = new AiChatMessageMapper();

    private AiChatMessage createMessage() {
        AiChatMessage message = new AiChatMessage();
        message.setId(1);

        AiChatSession session = new AiChatSession();
        session.setId(10);
        message.setSession(session);

        User user = new User();
        user.setId(5);
        message.setUser(user);

        message.setRole(AiChatMessage.Role.user);
        message.setContent("テストメッセージ");
        message.setCreatedAt(Timestamp.valueOf("2025-01-01 12:00:00"));
        return message;
    }

    @Test
    @DisplayName("エンティティからDTOに正しく変換できる")
    void toDtoConvertsCorrectly() {
        AiChatMessage message = createMessage();

        AiChatMessageResponseDto dto = mapper.toDto(message);

        assertThat(dto.id()).isEqualTo(1);
        assertThat(dto.sessionId()).isEqualTo(10);
        assertThat(dto.userId()).isEqualTo(5);
        assertThat(dto.role()).isEqualTo("user");
        assertThat(dto.content()).isEqualTo("テストメッセージ");
        assertThat(dto.createdAt()).isEqualTo(Timestamp.valueOf("2025-01-01 12:00:00"));
    }

    @Test
    @DisplayName("assistantロールが正しく変換される")
    void toDtoWithAssistantRole() {
        AiChatMessage message = createMessage();
        message.setRole(AiChatMessage.Role.assistant);

        AiChatMessageResponseDto dto = mapper.toDto(message);

        assertThat(dto.role()).isEqualTo("assistant");
    }

    @Test
    @DisplayName("nullを渡すとNullPointerExceptionを投げる")
    void toDtoThrowsOnNull() {
        assertThatThrownBy(() -> mapper.toDto(null))
            .isInstanceOf(NullPointerException.class)
            .hasMessageContaining("null");
    }

    @Test
    @DisplayName("異なるsessionId・userIdで正しく変換される")
    void toDtoWithDifferentIds() {
        AiChatMessage message = createMessage();
        message.getSession().setId(99);
        message.getUser().setId(42);

        AiChatMessageResponseDto dto = mapper.toDto(message);

        assertThat(dto.sessionId()).isEqualTo(99);
        assertThat(dto.userId()).isEqualTo(42);
    }

    @Test
    @DisplayName("空のコンテンツでも正しく変換される")
    void toDtoWithEmptyContent() {
        AiChatMessage message = createMessage();
        message.setContent("");

        AiChatMessageResponseDto dto = mapper.toDto(message);

        assertThat(dto.content()).isEmpty();
        assertThat(dto.id()).isEqualTo(1);
    }
}
