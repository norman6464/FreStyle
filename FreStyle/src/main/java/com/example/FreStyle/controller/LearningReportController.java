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

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GenerateMonthlyReportUseCase;
import com.example.FreStyle.usecase.GetMonthlyReportUseCase;
import com.example.FreStyle.usecase.GetReportListUseCase;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotNull;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/reports")
@Slf4j
public class LearningReportController {

    private final GenerateMonthlyReportUseCase generateMonthlyReportUseCase;
    private final GetMonthlyReportUseCase getMonthlyReportUseCase;
    private final GetReportListUseCase getReportListUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<List<LearningReportDto>> getReportList(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        List<LearningReportDto> reports = getReportListUseCase.execute(user.getId());
        return ResponseEntity.ok(reports);
    }

    @GetMapping("/{year}/{month}")
    public ResponseEntity<LearningReportDto> getMonthlyReport(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer year,
            @PathVariable Integer month) {
        User user = resolveUser(jwt);
        LearningReportDto report = getMonthlyReportUseCase.execute(user.getId(), year, month);
        if (report == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(report);
    }

    @PostMapping("/generate")
    public ResponseEntity<LearningReportDto> generateReport(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody GenerateReportRequest request) {
        User user = resolveUser(jwt);
        LearningReportDto report = generateMonthlyReportUseCase.execute(user, request.year(), request.month());
        return ResponseEntity.ok(report);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }

    record GenerateReportRequest(@NotNull Integer year, @NotNull Integer month) {}
}
