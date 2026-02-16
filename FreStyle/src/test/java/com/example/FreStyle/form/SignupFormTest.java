package com.example.FreStyle.form;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class SignupFormTest {

    private static Validator validator;

    @BeforeAll
    static void setUp() {
        validator = Validation.buildDefaultValidatorFactory().getValidator();
    }

    @Test
    void 正常なフォームはバリデーションエラーなし() {
        SignupForm form = new SignupForm("user@example.com", "password123", "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void メールアドレスが空の場合エラー() {
        SignupForm form = new SignupForm("", "password123", "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスを入力してください"));
    }

    @Test
    void メールアドレスの形式が不正な場合エラー() {
        SignupForm form = new SignupForm("not-email", "password123", "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスは正しい形式で入力してください。"));
    }

    @Test
    void パスワードが空の場合エラー() {
        SignupForm form = new SignupForm("user@example.com", "", "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("パスワードを入力してください"));
    }

    @Test
    void パスワードが8文字未満の場合エラー() {
        SignupForm form = new SignupForm("user@example.com", "short", "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("正しい桁数を入力してください"));
    }

    @Test
    void ユーザー名が空の場合エラー() {
        SignupForm form = new SignupForm("user@example.com", "password123", "");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("ユーザー名を入力してください"));
    }

    @Test
    void ユーザー名がnullの場合エラー() {
        SignupForm form = new SignupForm("user@example.com", "password123", null);
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("ユーザー名を入力してください"));
    }

    @Test
    void パスワードがちょうど8文字の場合バリデーション通過() {
        SignupForm form = new SignupForm("user@example.com", "12345678", "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void メールアドレスがnullの場合エラー() {
        SignupForm form = new SignupForm(null, "password123", "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスを入力してください"));
    }

    @Test
    void パスワードがnullの場合エラー() {
        SignupForm form = new SignupForm("user@example.com", null, "テストユーザー");
        Set<ConstraintViolation<SignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("パスワードを入力してください"));
    }
}
