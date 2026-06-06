package com.normanblog.frestyle.dto;

/**
 * 教材作成・更新のリクエストボディ。
 *
 * <p>作成時は courseId 必須(service で検証)。更新時は courseId を無視する(所属コースは変えない)。
 * companyId / createdByUserId は所属コース・認証ユーザーから決める。
 */
public record TeachingMaterialRequest(
    Long courseId, String title, String content, Integer orderInCourse, Boolean isPublished) {}
