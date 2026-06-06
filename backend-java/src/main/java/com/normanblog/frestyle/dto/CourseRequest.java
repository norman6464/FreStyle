package com.normanblog.frestyle.dto;

/**
 * コース作成・更新のリクエストボディ。
 *
 * <p>companyId / createdByUserId はリクエストからは受け取らず、認証ユーザーから決める。
 * sortOrder / isPublished は未指定なら既定(100 / false)を service 側で補う。
 */
public record CourseRequest(
    String title, String description, Integer sortOrder, Boolean isPublished) {}
