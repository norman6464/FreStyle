package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.User;

import lombok.RequiredArgsConstructor;

/**
 * 練習セッション作成ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたシナリオIDに基づいて新しい練習セッションを作成</li>
 *   <li>シナリオ名からセッションタイトルを自動生成</li>
 *   <li>セッション種別を「practice」に設定</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>セッションタイトル形式: 「練習: {シナリオ名}」</li>
 *   <li>セッション種別は必ず「practice」</li>
 *   <li>シナリオIDが無効な場合はGetPracticeScenarioByIdUseCaseでエラー</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 *   <li>複数のサービスを組み合わせたビジネスフロー</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class CreatePracticeSessionUseCase {

    private final GetPracticeScenarioByIdUseCase getPracticeScenarioByIdUseCase;
    private final CreateAiChatSessionUseCase createAiChatSessionUseCase;

    /**
     * 練習セッションを作成
     *
     * @param user ログインユーザー
     * @param scenarioId 練習シナリオID
     * @return 作成された練習セッション情報（DTO）
     * @throws RuntimeException シナリオIDが無効な場合、またはセッション作成に失敗した場合
     */
    @Transactional
    public AiChatSessionDto execute(User user, Integer scenarioId) {
        // 1. シナリオ情報を取得
        PracticeScenarioDto scenario = getPracticeScenarioByIdUseCase.execute(scenarioId);

        // 2. セッションタイトルを生成
        String title = "練習: " + scenario.name();

        // 3. 練習セッションを作成
        // sessionType="practice", scenarioId=指定されたID
        return createAiChatSessionUseCase.execute(
                user.getId(),
                title,
                null,           // relatedRoomId: 練習モードでは不要
                null,           // scene: 練習モードでは不要
                "practice",     // sessionType: 練習モード
                scenarioId      // scenarioId: シナリオID
        );
    }
}
