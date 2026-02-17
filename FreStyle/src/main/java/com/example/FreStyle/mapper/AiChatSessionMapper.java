package com.example.FreStyle.mapper;

import java.util.Objects;

import org.springframework.stereotype.Component;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;

/**
 * AI Chatセッションのマッピングクラス
 *
 * <p>役割:</p>
 * <ul>
 *   <li>AiChatSessionエンティティ ⇔ AiChatSessionDTOの相互変換</li>
 *   <li>プレゼンテーション層とドメイン層の境界を明確化</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>プレゼンテーション層とアプリケーション層の間のマッピング層</li>
 *   <li>DTOとEntityの変換ロジックを一箇所に集約</li>
 * </ul>
 */
@Component
public class AiChatSessionMapper {

    /**
     * エンティティからDTOへ変換
     *
     * @param session AI Chatセッションエンティティ
     * @return AI ChatセッションDTO（APIレスポンス用）
     * @throws IllegalArgumentException sessionがnullの場合
     */
    public AiChatSessionDto toDto(AiChatSession session) {
        Objects.requireNonNull(session, "AiChatSessionエンティティがnullです");

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
