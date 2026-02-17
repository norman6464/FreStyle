package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.AxisAnalysisDto;
import com.example.FreStyle.dto.CommunicationScoreSummaryDto;
import com.example.FreStyle.dto.ScoreTrendDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetAxisAnalysisUseCase;
import com.example.FreStyle.usecase.GetCommunicationScoreSummaryUseCase;
import com.example.FreStyle.usecase.GetScoreTrendUseCase;

import jakarta.validation.constraints.Min;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/scores")
@Validated
@Slf4j
public class ScoreTrendController {

    private final GetScoreTrendUseCase getScoreTrendUseCase;
    private final GetAxisAnalysisUseCase getAxisAnalysisUseCase;
    private final GetCommunicationScoreSummaryUseCase getCommunicationScoreSummaryUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping("/trend")
    public ResponseEntity<ScoreTrendDto> getTrend(
            @AuthenticationPrincipal Jwt jwt,
            @RequestParam(defaultValue = "30") @Min(1) int days) {
        User user = resolveUser(jwt);
        log.info("GET /api/scores/trend - userId={}, days={}", user.getId(), days);
        ScoreTrendDto result = getScoreTrendUseCase.execute(user.getId(), days);
        return ResponseEntity.ok(result);
    }

    @GetMapping("/axis-analysis")
    public ResponseEntity<AxisAnalysisDto> getAxisAnalysis(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("GET /api/scores/axis-analysis - userId={}", user.getId());
        AxisAnalysisDto result = getAxisAnalysisUseCase.execute(user.getId());
        return ResponseEntity.ok(result);
    }

    @GetMapping("/summary")
    public ResponseEntity<CommunicationScoreSummaryDto> getSummary(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("GET /api/scores/summary - userId={}", user.getId());
        CommunicationScoreSummaryDto result = getCommunicationScoreSummaryUseCase.execute(user.getId());
        return ResponseEntity.ok(result);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }
}
