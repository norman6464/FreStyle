package com.normanblog.frestyle.dto;

/**
 * プロフィールアイコン用の S3 直接アップロード URL。Go 版の JSON 形と互換のフィールド名にしている。
 *
 * <p>{@code uploadUrl} は PUT 対象、{@code imageUrl} はアップロード後に表示する URL。
 */
public record ProfileImageUploadUrl(
    String uploadUrl, String imageUrl, String key, int expiresIn) {}
