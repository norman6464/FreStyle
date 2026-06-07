package com.normanblog.frestyle.dto;

/** GET/PUT /api/v2/company/settings のレスポンス。会社単位の設定。 */
public record CompanySettingsResponse(boolean aiChatEnabledForTrainees) {}
