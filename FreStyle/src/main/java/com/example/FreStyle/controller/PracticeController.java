package com.example.FreStyle.controller;

import java.util.List;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CreatePracticeSessionUseCase;
import com.example.FreStyle.usecase.GetAllPracticeScenariosUseCase;
import com.example.FreStyle.usecase.GetPracticeScenarioByIdUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/**
 * 練習モードAPIコントローラー
 *
 * <p>役割:</p>
 * <ul>
 *   <li>練習シナリオ関連のREST APIエンドポイントを提供</li>
 *   <li>リクエストの検証と認証チェック</li>
 *   <li>UseCaseを呼び出してビジネスロジックを実行</li>
 *   <li>レスポンスの整形とエラーハンドリング</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>プレゼンテーション層（外側の層）</li>
 *   <li>HTTPリクエスト/レスポンスの処理のみを担当</li>
 *   <li>ビジネスロジックはUseCaseに委譲</li>
 * </ul>
 *
 * <p>エンドポイント:</p>
 * <ul>
 *   <li>GET /api/practice/scenarios - シナリオ一覧取得</li>
 *   <li>GET /api/practice/scenarios/{scenarioId} - シナリオ詳細取得</li>
 *   <li>POST /api/practice/sessions - 練習セッション作成</li>
 * </ul>
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/api/practice")
@Slf4j
public class PracticeController {

    // UseCaseクラス（アプリケーション層）
    private final GetAllPracticeScenariosUseCase getAllPracticeScenariosUseCase;
    private final GetPracticeScenarioByIdUseCase getPracticeScenarioByIdUseCase;
    private final CreatePracticeSessionUseCase createPracticeSessionUseCase;

    // 認証サービス
    private final UserIdentityService userIdentityService;

    /**
     * 練習シナリオ一覧を取得
     *
     * <p>すべての練習シナリオを取得して返却します。</p>
     *
     * @param jwt JWTトークン（認証済みユーザー）
     * @return 練習シナリオDTOのリスト
     */
    @GetMapping("/scenarios")
    public ResponseEntity<List<PracticeScenarioDto>> getScenarios(@AuthenticationPrincipal Jwt jwt) {
        log.info("========== GET /api/practice/scenarios ==========");

        // 認証チェック
        resolveUser(jwt);

        // ユースケース実行
        List<PracticeScenarioDto> scenarios = getAllPracticeScenariosUseCase.execute();
        log.info("✅ シナリオ一覧取得成功 - 件数: {}", scenarios.size());

        return ResponseEntity.ok(scenarios);
    }

    /**
     * 練習シナリオ詳細を取得
     *
     * <p>指定されたIDの練習シナリオ情報を取得して返却します。</p>
     *
     * @param jwt JWTトークン（認証済みユーザー）
     * @param scenarioId 取得したいシナリオのID
     * @return 練習シナリオDTO
     */
    @GetMapping("/scenarios/{scenarioId}")
    public ResponseEntity<PracticeScenarioDto> getScenario(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer scenarioId
    ) {
        log.info("========== GET /api/practice/scenarios/{} ==========", scenarioId);

        // 認証チェック
        resolveUser(jwt);

        // ユースケース実行
        PracticeScenarioDto scenario = getPracticeScenarioByIdUseCase.execute(scenarioId);
        log.info("✅ シナリオ詳細取得成功");

        return ResponseEntity.ok(scenario);
    }

    /**
     * 練習セッションを作成
     *
     * <p>指定されたシナリオIDに基づいて新しい練習セッションを作成します。</p>
     *
     * <p>処理の流れ:</p>
     * <ol>
     *   <li>JWTトークンからユーザー情報を取得</li>
     *   <li>CreatePracticeSessionUseCaseを実行</li>
     *   <li>作成されたセッション情報を返却</li>
     * </ol>
     *
     * @param jwt JWTトークン（認証済みユーザー）
     * @param request リクエストボディ（scenarioIdを含む）
     * @return 作成された練習セッション情報
     */
    @PostMapping("/sessions")
    public ResponseEntity<AiChatSessionDto> createPracticeSession(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody CreatePracticeSessionRequest request
    ) {
        log.info("========== POST /api/practice/sessions ==========");

        // 認証チェック & ユーザー取得
        User user = resolveUser(jwt);

        // ユースケース実行
        AiChatSessionDto session = createPracticeSessionUseCase.execute(user, request.scenarioId());
        log.info("✅ 練習セッション作成成功 - sessionId: {}", session.getId());

        return ResponseEntity.ok(session);
    }

    /**
     * 練習セッション作成リクエスト
     *
     * @param scenarioId 練習シナリオID
     */
    record CreatePracticeSessionRequest(Integer scenarioId) {}

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }
}

