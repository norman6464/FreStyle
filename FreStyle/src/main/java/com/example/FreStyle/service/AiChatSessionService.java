package com.example.FreStyle.service;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AiChatSessionService {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final UserRepository userRepository;
    private final ChatRoomRepository chatRoomRepository;

    /**
     * 新しいセッションを作成
     */
    @Transactional
    public AiChatSessionDto createSession(Integer userId, String title, Integer relatedRoomId) {
        return createSession(userId, title, relatedRoomId, null);
    }

    /**
     * 新しいセッションを作成（シーン指定付き）
     */
    @Transactional
    public AiChatSessionDto createSession(Integer userId, String title, Integer relatedRoomId, String scene) {
        return createSession(userId, title, relatedRoomId, scene, "normal", null);
    }

    /**
     * 新しいセッションを作成（全パラメータ指定）
     */
    @Transactional
    public AiChatSessionDto createSession(Integer userId, String title, Integer relatedRoomId,
                                           String scene, String sessionType, Integer scenarioId) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("ユーザーが見つかりません: " + userId));

        AiChatSession session = new AiChatSession();
        session.setUser(user);
        session.setTitle(title);
        session.setScene(scene);
        session.setSessionType(sessionType);
        session.setScenarioId(scenarioId);

        // 関連ルームが指定されている場合は設定
        if (relatedRoomId != null) {
            ChatRoom relatedRoom = chatRoomRepository.findById(relatedRoomId)
                    .orElseThrow(() -> new RuntimeException("チャットルームが見つかりません: " + relatedRoomId));
            session.setRelatedRoom(relatedRoom);
        }

        AiChatSession saved = aiChatSessionRepository.save(session);
        return toDto(saved);
    }

    /**
     * 指定ユーザーのセッション一覧を取得
     */
    @Transactional(readOnly = true)
    public List<AiChatSessionDto> getSessionsByUserId(Integer userId) {
        List<AiChatSession> sessions = aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId);
        return sessions.stream()
                .map(this::toDto)
                .toList();
    }

    /**
     * セッションIDとユーザーIDでセッションを取得（権限チェック付き）
     */
    @Transactional(readOnly = true)
    public AiChatSessionDto getSessionByIdAndUserId(Integer sessionId, Integer userId) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new RuntimeException("セッションが見つかりません: " + sessionId));
        return toDto(session);
    }

    /**
     * セッションのタイトルを更新
     */
    @Transactional
    public AiChatSessionDto updateSessionTitle(Integer sessionId, Integer userId, String newTitle) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new RuntimeException("セッションが見つかりません: " + sessionId));
        session.setTitle(newTitle);
        AiChatSession saved = aiChatSessionRepository.save(session);
        return toDto(saved);
    }

    /**
     * セッションを削除
     */
    @Transactional
    public void deleteSession(Integer sessionId, Integer userId) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new RuntimeException("セッションが見つかりません: " + sessionId));
        aiChatSessionRepository.delete(session);
    }

    /**
     * 指定ルームに関連するセッション一覧を取得
     */
    @Transactional(readOnly = true)
    public List<AiChatSessionDto> getSessionsByRelatedRoomId(Integer roomId) {
        List<AiChatSession> sessions = aiChatSessionRepository.findByRelatedRoomId(roomId);
        return sessions.stream()
                .map(this::toDto)
                .toList();
    }

    /**
     * EntityからDTOへ変換
     */
    private AiChatSessionDto toDto(AiChatSession session) {
        return new AiChatSessionDto(
                session.getId(),
                session.getUser().getId(),
                session.getTitle(),
                session.getRelatedRoom() != null ? session.getRelatedRoom().getId() : null,
                session.getScene(),
                session.getSessionType(),
                session.getScenarioId(),
                session.getCreatedAt(),
                session.getUpdatedAt()
        );
    }

}
