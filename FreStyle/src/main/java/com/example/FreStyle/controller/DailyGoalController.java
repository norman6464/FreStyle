package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotNull;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.DailyGoalDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetTodayDailyGoalUseCase;
import com.example.FreStyle.usecase.IncrementDailyGoalUseCase;
import com.example.FreStyle.usecase.SetDailyGoalTargetUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/daily-goals")
@Slf4j
public class DailyGoalController {

    private final GetTodayDailyGoalUseCase getTodayDailyGoalUseCase;
    private final SetDailyGoalTargetUseCase setDailyGoalTargetUseCase;
    private final IncrementDailyGoalUseCase incrementDailyGoalUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping("/today")
    public ResponseEntity<DailyGoalDto> getToday(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("今日の日次目標取得: userId={}", user.getId());
        DailyGoalDto dto = getTodayDailyGoalUseCase.execute(user.getId());
        return ResponseEntity.ok(dto);
    }

    @PutMapping("/target")
    public ResponseEntity<Void> setTarget(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody SetTargetRequest request) {
        User user = resolveUser(jwt);
        log.info("日次目標設定: userId={}, target={}", user.getId(), request.target());
        setDailyGoalTargetUseCase.execute(user, request.target());
        return ResponseEntity.ok().build();
    }

    @PostMapping("/increment")
    public ResponseEntity<DailyGoalDto> increment(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("日次目標インクリメント: userId={}", user.getId());
        DailyGoalDto dto = incrementDailyGoalUseCase.execute(user);
        return ResponseEntity.ok(dto);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }

    record SetTargetRequest(@NotNull Integer target) {}
}
