package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.ScoreGoalDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetScoreGoalUseCase;
import com.example.FreStyle.usecase.SaveScoreGoalUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/score-goal")
@Slf4j
public class ScoreGoalController {

    private final GetScoreGoalUseCase getScoreGoalUseCase;
    private final SaveScoreGoalUseCase saveScoreGoalUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<ScoreGoalDto> getGoal(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("スコア目標取得: userId={}", user.getId());
        ScoreGoalDto dto = getScoreGoalUseCase.execute(user.getId());
        if (dto == null) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.ok(dto);
    }

    @PutMapping
    public ResponseEntity<Void> saveGoal(@AuthenticationPrincipal Jwt jwt, @RequestBody SaveGoalRequest request) {
        User user = resolveUser(jwt);
        log.info("スコア目標設定: userId={}, goalScore={}", user.getId(), request.goalScore());
        saveScoreGoalUseCase.execute(user, request.goalScore());
        return ResponseEntity.noContent().build();
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }

    public record SaveGoalRequest(Double goalScore) {}
}
