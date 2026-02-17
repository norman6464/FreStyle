package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.mapper.AiChatMessageMapper;
import com.example.FreStyle.repository.AiChatMessageRepository;

import lombok.RequiredArgsConstructor;

/**
 * ユーザー別AI Chatメッセージ一覧取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたユーザーIDのすべてのメッセージを取得</li>
 *   <li>作成日時の昇順でソート</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>メッセージは作成日時の古い順で返却</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class GetAiChatMessagesByUserIdUseCase {

    private final AiChatMessageRepository aiChatMessageRepository;
    private final AiChatMessageMapper mapper;

    /**
     * 指定ユーザーのメッセージ一覧を取得
     *
     * @param userId ユーザーID
     * @return メッセージDTOのリスト（作成日時昇順）
     */
    @Transactional(readOnly = true)
    public List<AiChatMessageResponseDto> execute(Integer userId) {
        List<AiChatMessage> messages = aiChatMessageRepository.findByUserIdOrderByCreatedAtAsc(userId);
        return messages.stream()
                .map(mapper::toDto)
                .toList();
    }
}
