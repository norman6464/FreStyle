package com.example.FreStyle.form;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class ProfileFormTest {

    private static Validator validator;

    @BeforeAll
    static void setUp() {
        validator = Validation.buildDefaultValidatorFactory().getValidator();
    }

    @Test
    void 正常なフォームはバリデーションエラーなし() {
        ProfileForm form = new ProfileForm("テストユーザー", "自己紹介文", null, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void 名前が空の場合エラー() {
        ProfileForm form = new ProfileForm("", "自己紹介文", null, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("ユーザー名を入力してください"));
    }

    @Test
    void 名前がnullの場合エラー() {
        ProfileForm form = new ProfileForm(null, "自己紹介文", null, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("ユーザー名を入力してください"));
    }

    @Test
    void bioがnullでもバリデーション通過() {
        ProfileForm form = new ProfileForm("テストユーザー", null, null, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void iconUrlがnullでもバリデーション通過() {
        ProfileForm form = new ProfileForm("テストユーザー", "bio", null, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void iconUrlが設定されている場合もバリデーション通過() {
        ProfileForm form = new ProfileForm("テストユーザー", "bio", "https://example.com/icon.jpg", null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void ステータスが100文字以内なら通過() {
        ProfileForm form = new ProfileForm("テストユーザー", "bio", null, "学習中");
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void ステータスが100文字を超えるとエラー() {
        String longStatus = "あ".repeat(101);
        ProfileForm form = new ProfileForm("テストユーザー", "bio", null, longStatus);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("ステータスは100文字以内で入力してください"));
    }

    @Test
    void bioが500文字以内なら通過() {
        String bio = "あ".repeat(500);
        ProfileForm form = new ProfileForm("テストユーザー", bio, null, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void bioが500文字を超えるとエラー() {
        String longBio = "あ".repeat(501);
        ProfileForm form = new ProfileForm("テストユーザー", longBio, null, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("自己紹介は500文字以内で入力してください"));
    }

    @Test
    void iconUrlが2048文字を超えるとエラー() {
        String longUrl = "https://example.com/" + "a".repeat(2030);
        ProfileForm form = new ProfileForm("テストユーザー", "bio", longUrl, null);
        Set<ConstraintViolation<ProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("アイコンURLが長すぎます"));
    }
}
