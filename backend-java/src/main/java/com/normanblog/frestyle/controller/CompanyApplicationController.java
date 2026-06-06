package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.CompanyApplicationResponse;
import com.normanblog.frestyle.dto.CreateCompanyApplicationRequest;
import com.normanblog.frestyle.service.CompanyApplicationService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 企業利用申請の公開エンドポイント(ログイン前の未登録ユーザーが使う)。 */
@RestController
@RequestMapping("/api/v2/company-applications")
public class CompanyApplicationController {

  private final CompanyApplicationService applications;

  public CompanyApplicationController(CompanyApplicationService applications) {
    this.applications = applications;
  }

  /** 会社名 / 氏名 / メール / 任意メッセージで利用申請を受け付ける(認証不要)。 */
  @PostMapping
  public ResponseEntity<CompanyApplicationResponse> create(
      @Valid @RequestBody CreateCompanyApplicationRequest request) {
    CompanyApplicationResponse body =
        CompanyApplicationResponse.from(
            applications.create(
                request.companyName(),
                request.applicantName(),
                request.email(),
                request.message()));

    return ResponseEntity.status(HttpStatus.CREATED).body(body);
  }
}
