package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.GenerateReportRequest;
import com.normanblog.frestyle.dto.LearningReportResponse;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.LearningReportService;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 学習レポート API。userId は IDOR 対策で受け取らず、常に認証ユーザー自身のレポートを扱う。
 */
@RestController
@RequestMapping("/api/v2/learning-reports")
public class LearningReportController {

  private final LearningReportService reports;
  private final CurrentUserProvider currentUser;

  public LearningReportController(
      LearningReportService reports, CurrentUserProvider currentUser) {
    this.reports = reports;
    this.currentUser = currentUser;
  }

  @GetMapping
  public List<LearningReportResponse> list() {
    Long userId = currentUser.require().getId();
    return reports.list(userId).stream().map(LearningReportResponse::from).toList();
  }

  /** 指定月のレポート生成を受け付け、202 Accepted で pending のレポートを返す。 */
  @PostMapping("/generate")
  public ResponseEntity<LearningReportResponse> generate(
      @Valid @RequestBody GenerateReportRequest request) {
    Long userId = currentUser.require().getId();
    LearningReportResponse body =
        LearningReportResponse.from(
            reports.requestMonthly(userId, request.year(), request.month()));

    return ResponseEntity.status(HttpStatus.ACCEPTED).body(body);
  }
}
