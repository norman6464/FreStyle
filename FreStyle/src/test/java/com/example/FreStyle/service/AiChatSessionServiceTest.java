package com.example.FreStyle.service;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.UserRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.sql.Timestamp;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AiChatSessionServiceTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private UserRepository userRepository;

    @Mock
    private ChatRoomRepository chatRoomRepository;

    @InjectMocks
    private AiChatSessionService aiChatSessionService;

    private User createUser(Integer id) {
        User user = new User();
        user.setId(id);
        return user;
    }

    private AiChatSession createSession(Integer id, User user, String title) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        session.setUser(user);
        session.setTitle(title);
        session.setSessionType("normal");
        session.setCreatedAt(new Timestamp(System.currentTimeMillis()));
        session.setUpdatedAt(new Timestamp(System.currentTimeMillis()));
        return session;
    }

    @Test
    @DisplayName("createSession: 新しいセッションを作成してDTOを返す")
    void createSession_createsAndReturnsDto() {
        User user = createUser(1);
        when(userRepository.findById(1)).thenReturn(Optional.of(user));

        AiChatSession saved = createSession(10, user, "テストセッション");
        when(aiChatSessionRepository.save(any())).thenReturn(saved);

        AiChatSessionDto dto = aiChatSessionService.createSession(1, "テストセッション", null);

        assertEquals(10, dto.id());
        assertEquals(1, dto.userId());
        assertEquals("テストセッション", dto.title());

        ArgumentCaptor<AiChatSession> captor = ArgumentCaptor.forClass(AiChatSession.class);
        verify(aiChatSessionRepository).save(captor.capture());
        assertEquals(user, captor.getValue().getUser());
        assertEquals("テストセッション", captor.getValue().getTitle());
    }

    @Test
    @DisplayName("createSession: 関連ルーム付きでセッションを作成")
    void createSession_withRelatedRoom() {
        User user = createUser(1);
        ChatRoom room = new ChatRoom();
        room.setId(5);
        when(userRepository.findById(1)).thenReturn(Optional.of(user));
        when(chatRoomRepository.findById(5)).thenReturn(Optional.of(room));

        AiChatSession saved = createSession(11, user, "ルーム付き");
        saved.setRelatedRoom(room);
        when(aiChatSessionRepository.save(any())).thenReturn(saved);

        AiChatSessionDto dto = aiChatSessionService.createSession(1, "ルーム付き", 5, "meeting", "normal", null);

        assertEquals(5, dto.relatedRoomId());
    }

    @Test
    @DisplayName("createSession: ユーザーが存在しない場合に例外")
    void createSession_throwsWhenUserNotFound() {
        when(userRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatSessionService.createSession(999, "テスト", null));
        assertTrue(ex.getMessage().contains("ユーザーが見つかりません"));
    }

    @Test
    @DisplayName("getSessionsByUserId: ユーザーのセッション一覧を返す")
    void getSessionsByUserId_returnsList() {
        User user = createUser(1);
        AiChatSession s1 = createSession(1, user, "セッション1");
        AiChatSession s2 = createSession(2, user, "セッション2");
        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(s1, s2));

        List<AiChatSessionDto> result = aiChatSessionService.getSessionsByUserId(1);

        assertEquals(2, result.size());
        assertEquals("セッション1", result.get(0).title());
        assertEquals("セッション2", result.get(1).title());
    }

    @Test
    @DisplayName("getSessionByIdAndUserId: セッションを返す")
    void getSessionByIdAndUserId_returnsSession() {
        User user = createUser(1);
        AiChatSession session = createSession(10, user, "取得テスト");
        when(aiChatSessionRepository.findByIdAndUserId(10, 1))
                .thenReturn(Optional.of(session));

        AiChatSessionDto dto = aiChatSessionService.getSessionByIdAndUserId(10, 1);

        assertEquals("取得テスト", dto.title());
    }

    @Test
    @DisplayName("getSessionByIdAndUserId: 存在しない場合に例外")
    void getSessionByIdAndUserId_throwsWhenNotFound() {
        when(aiChatSessionRepository.findByIdAndUserId(999, 1))
                .thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatSessionService.getSessionByIdAndUserId(999, 1));
        assertTrue(ex.getMessage().contains("セッションが見つかりません"));
    }

    @Test
    @DisplayName("updateSessionTitle: タイトルを更新してDTOを返す")
    void updateSessionTitle_updatesAndReturnsDto() {
        User user = createUser(1);
        AiChatSession session = createSession(10, user, "旧タイトル");
        when(aiChatSessionRepository.findByIdAndUserId(10, 1))
                .thenReturn(Optional.of(session));
        when(aiChatSessionRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));

        AiChatSessionDto dto = aiChatSessionService.updateSessionTitle(10, 1, "新タイトル");

        assertEquals("新タイトル", dto.title());
    }

    @Test
    @DisplayName("deleteSession: セッションを削除する")
    void deleteSession_deletesSession() {
        User user = createUser(1);
        AiChatSession session = createSession(10, user, "削除対象");
        when(aiChatSessionRepository.findByIdAndUserId(10, 1))
                .thenReturn(Optional.of(session));

        aiChatSessionService.deleteSession(10, 1);

        verify(aiChatSessionRepository).delete(session);
    }

    @Test
    @DisplayName("getSessionsByRelatedRoomId: 関連ルームのセッション一覧を返す")
    void getSessionsByRelatedRoomId_returnsList() {
        User user = createUser(1);
        ChatRoom room = new ChatRoom();
        room.setId(5);
        AiChatSession s1 = createSession(1, user, "ルーム関連1");
        s1.setRelatedRoom(room);
        AiChatSession s2 = createSession(2, user, "ルーム関連2");
        s2.setRelatedRoom(room);
        when(aiChatSessionRepository.findByRelatedRoomId(5))
                .thenReturn(List.of(s1, s2));

        List<AiChatSessionDto> result = aiChatSessionService.getSessionsByRelatedRoomId(5);

        assertEquals(2, result.size());
        assertEquals("ルーム関連1", result.get(0).title());
        assertEquals(5, result.get(0).relatedRoomId());
    }

    @Test
    @DisplayName("getSessionsByRelatedRoomId: セッションがない場合は空リスト")
    void getSessionsByRelatedRoomId_returnsEmptyList() {
        when(aiChatSessionRepository.findByRelatedRoomId(999))
                .thenReturn(List.of());

        List<AiChatSessionDto> result = aiChatSessionService.getSessionsByRelatedRoomId(999);

        assertTrue(result.isEmpty());
    }

    @Test
    @DisplayName("createSession: 関連ルームが見つからない場合に例外")
    void createSession_throwsWhenRelatedRoomNotFound() {
        User user = createUser(1);
        when(userRepository.findById(1)).thenReturn(Optional.of(user));
        when(chatRoomRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatSessionService.createSession(1, "テスト", 999, "meeting", "normal", null));
        assertTrue(ex.getMessage().contains("チャットルームが見つかりません"));
    }

    @Test
    @DisplayName("updateSessionTitle: 存在しないセッションの場合に例外")
    void updateSessionTitle_throwsWhenNotFound() {
        when(aiChatSessionRepository.findByIdAndUserId(999, 1))
                .thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatSessionService.updateSessionTitle(999, 1, "新タイトル"));
        assertTrue(ex.getMessage().contains("セッションが見つかりません"));
    }

    @Test
    @DisplayName("deleteSession: 存在しないセッションの場合に例外")
    void deleteSession_throwsWhenNotFound() {
        when(aiChatSessionRepository.findByIdAndUserId(999, 1))
                .thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatSessionService.deleteSession(999, 1));
        assertTrue(ex.getMessage().contains("セッションが見つかりません"));
    }

    @Test
    @DisplayName("createSession: 3パラメータオーバーロードでscene=null, sessionType=normalが設定される")
    void createSession_threeParamOverload_setsDefaults() {
        User user = createUser(1);
        when(userRepository.findById(1)).thenReturn(Optional.of(user));

        AiChatSession saved = createSession(10, user, "デフォルト確認");
        when(aiChatSessionRepository.save(any())).thenReturn(saved);

        aiChatSessionService.createSession(1, "デフォルト確認", null);

        ArgumentCaptor<AiChatSession> captor = ArgumentCaptor.forClass(AiChatSession.class);
        verify(aiChatSessionRepository).save(captor.capture());
        assertNull(captor.getValue().getScene());
        assertEquals("normal", captor.getValue().getSessionType());
        assertNull(captor.getValue().getScenarioId());
    }
}
