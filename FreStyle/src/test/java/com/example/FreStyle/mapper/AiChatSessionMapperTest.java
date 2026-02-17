package com.example.FreStyle.mapper;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import java.sql.Timestamp;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;

@DisplayName("AiChatSessionMapper")
class AiChatSessionMapperTest {

    private final AiChatSessionMapper mapper = new AiChatSessionMapper();

    private AiChatSession createSession() {
        AiChatSession session = new AiChatSession();
        session.setId(1);

        User user = new User();
        user.setId(10);
        session.setUser(user);

        session.setTitle("テストセッション");
        session.setScene("meeting");
        session.setSessionType("normal");
        session.setScenarioId(null);
        session.setCreatedAt(Timestamp.valueOf("2025-01-01 12:00:00"));
        session.setUpdatedAt(Timestamp.valueOf("2025-01-01 13:00:00"));
        return session;
    }

    @Test
    @DisplayName("エンティティからDTOに正しく変換できる")
    void toDtoConvertsCorrectly() {
        AiChatSession session = createSession();

        AiChatSessionDto dto = mapper.toDto(session);

        assertThat(dto.id()).isEqualTo(1);
        assertThat(dto.userId()).isEqualTo(10);
        assertThat(dto.title()).isEqualTo("テストセッション");
        assertThat(dto.scene()).isEqualTo("meeting");
        assertThat(dto.sessionType()).isEqualTo("normal");
        assertThat(dto.relatedRoomId()).isNull();
    }

    @Test
    @DisplayName("relatedRoomがある場合はIDが設定される")
    void toDtoWithRelatedRoom() {
        AiChatSession session = createSession();
        ChatRoom room = new ChatRoom();
        room.setId(99);
        session.setRelatedRoom(room);

        AiChatSessionDto dto = mapper.toDto(session);

        assertThat(dto.relatedRoomId()).isEqualTo(99);
    }

    @Test
    @DisplayName("relatedRoomがnullの場合はrelatedRoomIdがnull")
    void toDtoWithoutRelatedRoom() {
        AiChatSession session = createSession();
        session.setRelatedRoom(null);

        AiChatSessionDto dto = mapper.toDto(session);

        assertThat(dto.relatedRoomId()).isNull();
    }

    @Test
    @DisplayName("nullを渡すとIllegalArgumentExceptionを投げる")
    void toDtoThrowsOnNull() {
        assertThatThrownBy(() -> mapper.toDto(null))
            .isInstanceOf(IllegalArgumentException.class)
            .hasMessageContaining("null");
    }
}
