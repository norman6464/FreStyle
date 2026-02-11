package com.example.FreStyle.controller;

import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
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
import com.example.FreStyle.service.AiChatSessionService;
import com.example.FreStyle.service.PracticeScenarioService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/practice")
public class PracticeController {

    private static final Logger logger = LoggerFactory.getLogger(PracticeController.class);
    private final PracticeScenarioService practiceScenarioService;
    private final AiChatSessionService aiChatSessionService;
    private final UserIdentityService userIdentityService;

    /**
     * シナリオ一覧を取得
     */
    @GetMapping("/scenarios")
    public ResponseEntity<List<PracticeScenarioDto>> getScenarios(@AuthenticationPrincipal Jwt jwt) {
        logger.info("========== GET /api/practice/scenarios ==========");

        try {
            String sub = jwt.getSubject();
            userIdentityService.findUserBySub(sub); // 認証チェック

            List<PracticeScenarioDto> scenarios = practiceScenarioService.getAllScenarios();
            logger.info("✅ シナリオ一覧取得成功 - 件数: {}", scenarios.size());

            return ResponseEntity.ok(scenarios);
        } catch (Exception e) {
            logger.error("❌ シナリオ一覧取得エラー: {}", e.getMessage());
            return ResponseEntity.internalServerError().build();
        }
    }

    /**
     * シナリオ詳細を取得
     */
    @GetMapping("/scenarios/{scenarioId}")
    public ResponseEntity<PracticeScenarioDto> getScenario(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer scenarioId
    ) {
        logger.info("========== GET /api/practice/scenarios/{} ==========", scenarioId);

        try {
            String sub = jwt.getSubject();
            userIdentityService.findUserBySub(sub);

            PracticeScenarioDto scenario = practiceScenarioService.getScenarioById(scenarioId);
            logger.info("✅ シナリオ詳細取得成功");

            return ResponseEntity.ok(scenario);
        } catch (RuntimeException e) {
            logger.error("❌ シナリオ詳細取得エラー: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }

    /**
     * 練習セッションを作成
     */
    @PostMapping("/sessions")
    public ResponseEntity<AiChatSessionDto> createPracticeSession(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody CreatePracticeSessionRequest request
    ) {
        logger.info("========== POST /api/practice/sessions ==========");

        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);

            PracticeScenarioDto scenario = practiceScenarioService.getScenarioById(request.scenarioId());
            String title = "練習: " + scenario.getName();

            AiChatSessionDto session = aiChatSessionService.createSession(
                    user.getId(), title, null, null, "practice", request.scenarioId());
            logger.info("✅ 練習セッション作成成功 - sessionId: {}", session.getId());

            return ResponseEntity.ok(session);
        } catch (Exception e) {
            logger.error("❌ 練習セッション作成エラー: {}", e.getMessage());
            return ResponseEntity.internalServerError().build();
        }
    }

    record CreatePracticeSessionRequest(Integer scenarioId) {}
}
