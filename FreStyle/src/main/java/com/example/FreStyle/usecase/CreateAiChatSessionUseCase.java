package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;

/**
 * AI Chatセッション作成ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>新しいAI Chatセッションを作成</li>
 *   <li>ユーザーとルームの存在確認</li>
 *   <li>セッションタイプ、シーン、関連ルームの設定</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>ユーザーIDが無効な場合はエラー</li>
 *   <li>関連ルームIDが指定されている場合は、ルームの存在確認を行う</li>
 *   <li>セッションタイプのデフォルトは「normal」</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 *   <li>複数のリポジトリを組み合わせたビジネスフロー</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class CreateAiChatSessionUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final UserRepository userRepository;
    private final ChatRoomRepository chatRoomRepository;
    private final AiChatSessionMapper mapper;

    /**
     * AI Chatセッションを作成（基本形）
     *
     * @param userId ユーザーID
     * @param title セッションタイトル
     * @param relatedRoomId 関連するチャットルームID（任意）
     * @return 作成されたセッション情報（DTO）
     * @throws ResourceNotFoundException ユーザーまたはルームが見つからない場合
     */
    @Transactional
    public AiChatSessionDto execute(Integer userId, String title, Integer relatedRoomId) {
        return execute(userId, title, relatedRoomId, null, "normal", null);
    }

    /**
     * AI Chatセッションを作成（シーン指定付き）
     *
     * @param userId ユーザーID
     * @param title セッションタイトル
     * @param relatedRoomId 関連するチャットルームID（任意）
     * @param scene シーン（任意）
     * @return 作成されたセッション情報（DTO）
     * @throws ResourceNotFoundException ユーザーまたはルームが見つからない場合
     */
    @Transactional
    public AiChatSessionDto execute(Integer userId, String title, Integer relatedRoomId, String scene) {
        return execute(userId, title, relatedRoomId, scene, "normal", null);
    }

    /**
     * AI Chatセッションを作成（全パラメータ指定）
     *
     * @param userId ユーザーID
     * @param title セッションタイトル
     * @param relatedRoomId 関連するチャットルームID（任意）
     * @param scene シーン（任意）
     * @param sessionType セッションタイプ（"normal", "practice"等）
     * @param scenarioId 練習シナリオID（練習モードの場合のみ）
     * @return 作成されたセッション情報（DTO）
     * @throws ResourceNotFoundException ユーザーまたはルームが見つからない場合
     */
    @Transactional
    public AiChatSessionDto execute(
            Integer userId,
            String title,
            Integer relatedRoomId,
            String scene,
            String sessionType,
            Integer scenarioId
    ) {
        // 1. ユーザーを取得
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new ResourceNotFoundException("ユーザーが見つかりません: ID=" + userId));

        // 2. セッションを作成
        AiChatSession session = new AiChatSession();
        session.setUser(user);
        session.setTitle(title);
        session.setScene(scene);
        session.setSessionType(sessionType);
        session.setScenarioId(scenarioId);

        // 3. 関連ルームが指定されている場合は設定
        if (relatedRoomId != null) {
            ChatRoom relatedRoom = chatRoomRepository.findById(relatedRoomId)
                    .orElseThrow(() -> new ResourceNotFoundException("チャットルームが見つかりません: ID=" + relatedRoomId));
            session.setRelatedRoom(relatedRoom);
        }

        // 4. 保存してDTOに変換
        AiChatSession saved = aiChatSessionRepository.save(session);
        return mapper.toDto(saved);
    }
}
