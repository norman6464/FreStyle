package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.http.ResponseEntity;
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

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/daily-goals")
public class DailyGoalController {

    private final GetTodayDailyGoalUseCase getTodayDailyGoalUseCase;
    private final SetDailyGoalTargetUseCase setDailyGoalTargetUseCase;
    private final IncrementDailyGoalUseCase incrementDailyGoalUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping("/today")
    public ResponseEntity<DailyGoalDto> getToday(@AuthenticationPrincipal Jwt jwt) {
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        DailyGoalDto dto = getTodayDailyGoalUseCase.execute(user.getId());
        return ResponseEntity.ok(dto);
    }

    @PutMapping("/target")
    public ResponseEntity<Void> setTarget(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody Map<String, Integer> body) {
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        setDailyGoalTargetUseCase.execute(user, body.get("target"));
        return ResponseEntity.ok().build();
    }

    @PostMapping("/increment")
    public ResponseEntity<DailyGoalDto> increment(@AuthenticationPrincipal Jwt jwt) {
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        DailyGoalDto dto = incrementDailyGoalUseCase.execute(user);
        return ResponseEntity.ok(dto);
    }
}
