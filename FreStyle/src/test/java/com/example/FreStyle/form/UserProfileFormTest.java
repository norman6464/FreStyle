package com.example.FreStyle.form;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class UserProfileFormTest {

    private static Validator validator;

    @BeforeAll
    static void setUp() {
        validator = Validation.buildDefaultValidatorFactory().getValidator();
    }

    @Test
    void 正常なフォームはバリデーションエラーなし() {
        UserProfileForm form = new UserProfileForm(
                "表示名", "自己紹介", "フレンドリー",
                List.of("積極的", "論理的"), "目標", "悩み", "具体的"
        );
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void 全てnullでもバリデーション通過() {
        UserProfileForm form = new UserProfileForm(null, null, null, null, null, null, null);
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void 表示名が100文字超の場合エラー() {
        String longName = "あ".repeat(101);
        UserProfileForm form = new UserProfileForm(longName, null, null, null, null, null, null);
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("表示名は100文字以内で入力してください"));
    }

    @Test
    void 表示名がちょうど100文字の場合バリデーション通過() {
        String name = "あ".repeat(100);
        UserProfileForm form = new UserProfileForm(name, null, null, null, null, null, null);
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void コミュニケーションスタイルが50文字超の場合エラー() {
        String longStyle = "あ".repeat(51);
        UserProfileForm form = new UserProfileForm(null, null, longStyle, null, null, null, null);
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("コミュニケーションスタイルは50文字以内で入力してください"));
    }

    @Test
    void コミュニケーションスタイルがちょうど50文字の場合バリデーション通過() {
        String style = "あ".repeat(50);
        UserProfileForm form = new UserProfileForm(null, null, style, null, null, null, null);
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void フィードバックスタイルが50文字超の場合エラー() {
        String longFeedback = "あ".repeat(51);
        UserProfileForm form = new UserProfileForm(null, null, null, null, null, null, longFeedback);
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("フィードバックスタイルは50文字以内で入力してください"));
    }

    @Test
    void フィードバックスタイルがちょうど50文字の場合バリデーション通過() {
        String feedback = "あ".repeat(50);
        UserProfileForm form = new UserProfileForm(null, null, null, null, null, null, feedback);
        Set<ConstraintViolation<UserProfileForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }
}
