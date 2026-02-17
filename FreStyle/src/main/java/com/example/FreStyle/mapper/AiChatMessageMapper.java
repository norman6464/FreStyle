package com.example.FreStyle.mapper;

import java.util.Objects;

import org.springframework.stereotype.Component;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatMessage;

/**
 * AI Chatメッセージのマッピングクラス
 *
 * <p>役割:</p>
 * <ul>
 *   <li>AiChatMessageエンティティ ⇔ AiChatMessageResponseDTOの相互変換</li>
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
public class AiChatMessageMapper {

    /**
     * エンティティからDTOへ変換
     *
     * @param message AI Chatメッセージエンティティ
     * @return AI ChatメッセージDTO（APIレスポンス用）
     * @throws NullPointerException messageがnullの場合
     */
    public AiChatMessageResponseDto toDto(AiChatMessage message) {
        Objects.requireNonNull(message, "AiChatMessageエンティティがnullです");

        return new AiChatMessageResponseDto(
            message.getId(),
            message.getSession().getId(),
            message.getUser().getId(),
            message.getRole().name(),
            message.getContent(),
            message.getCreatedAt()
        );
    }
}
