package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.UserRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("CreateAiChatSessionUseCase")
class CreateAiChatSessionUseCaseTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private UserRepository userRepository;

    @Mock
    private ChatRoomRepository chatRoomRepository;

    @Mock
    private AiChatSessionMapper mapper;

    @InjectMocks
    private CreateAiChatSessionUseCase useCase;

    @Nested
    @DisplayName("正常系")
    class Success {

        @Test
        @DisplayName("セッションを作成してDTOを返す")
        void shouldCreateSessionAndReturnDto() {
            User user = new User();
            user.setId(1);
            when(userRepository.findById(1)).thenReturn(Optional.of(user));

            AiChatSession savedSession = new AiChatSession();
            savedSession.setId(10);
            when(aiChatSessionRepository.save(any(AiChatSession.class))).thenReturn(savedSession);

            AiChatSessionDto expectedDto = new AiChatSessionDto();
            expectedDto.setId(10);
            when(mapper.toDto(savedSession)).thenReturn(expectedDto);

            AiChatSessionDto result = useCase.execute(1, "テストセッション", null);

            assertThat(result.getId()).isEqualTo(10);
            verify(aiChatSessionRepository).save(any(AiChatSession.class));
        }

        @Test
        @DisplayName("関連ルーム指定でセッションを作成する")
        void shouldCreateSessionWithRelatedRoom() {
            User user = new User();
            user.setId(1);
            when(userRepository.findById(1)).thenReturn(Optional.of(user));

            ChatRoom room = new ChatRoom();
            room.setId(5);
            when(chatRoomRepository.findById(5)).thenReturn(Optional.of(room));

            AiChatSession savedSession = new AiChatSession();
            savedSession.setId(10);
            when(aiChatSessionRepository.save(any(AiChatSession.class))).thenReturn(savedSession);

            AiChatSessionDto expectedDto = new AiChatSessionDto();
            expectedDto.setId(10);
            when(mapper.toDto(savedSession)).thenReturn(expectedDto);

            AiChatSessionDto result = useCase.execute(1, "テスト", 5);

            assertThat(result.getId()).isEqualTo(10);
            verify(chatRoomRepository).findById(5);
        }

        @Test
        @DisplayName("シーンとセッションタイプ指定でセッションを作成する")
        void shouldCreateSessionWithSceneAndType() {
            User user = new User();
            user.setId(1);
            when(userRepository.findById(1)).thenReturn(Optional.of(user));

            AiChatSession savedSession = new AiChatSession();
            savedSession.setId(10);
            when(aiChatSessionRepository.save(any(AiChatSession.class))).thenReturn(savedSession);

            AiChatSessionDto expectedDto = new AiChatSessionDto();
            expectedDto.setId(10);
            when(mapper.toDto(savedSession)).thenReturn(expectedDto);

            AiChatSessionDto result = useCase.execute(1, "練習", null, "meeting", "practice", 3);

            assertThat(result.getId()).isEqualTo(10);
        }
    }

    @Nested
    @DisplayName("異常系")
    class Error {

        @Test
        @DisplayName("ユーザーが見つからない場合はResourceNotFoundExceptionをスローする")
        void shouldThrowWhenUserNotFound() {
            when(userRepository.findById(999)).thenReturn(Optional.empty());

            assertThatThrownBy(() -> useCase.execute(999, "テスト", null))
                    .isInstanceOf(ResourceNotFoundException.class);
        }
    }
}
