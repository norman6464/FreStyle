package com.example.FreStyle.form;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class LoginFormTest {

    private static Validator validator;

    @BeforeAll
    static void setUp() {
        validator = Validation.buildDefaultValidatorFactory().getValidator();
    }

    @Test
    void 正常なフォームはバリデーションエラーなし() {
        LoginForm form = new LoginForm("user@example.com", "password123");
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void メールアドレスが空の場合エラー() {
        LoginForm form = new LoginForm("", "password123");
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスを入力してください"));
    }

    @Test
    void メールアドレスがnullの場合エラー() {
        LoginForm form = new LoginForm(null, "password123");
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスを入力してください"));
    }

    @Test
    void メールアドレスの形式が不正な場合エラー() {
        LoginForm form = new LoginForm("invalid-email", "password123");
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスは正しい形式で入力してください。"));
    }

    @Test
    void パスワードが空の場合エラー() {
        LoginForm form = new LoginForm("user@example.com", "");
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("パスワードを入力してください"));
    }

    @Test
    void パスワードがnullの場合エラー() {
        LoginForm form = new LoginForm("user@example.com", null);
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("パスワードを入力してください"));
    }

    @Test
    void パスワードが8文字未満の場合エラー() {
        LoginForm form = new LoginForm("user@example.com", "short");
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("正しい桁数を入力してください"));
    }

    @Test
    void パスワードがちょうど8文字の場合バリデーション通過() {
        LoginForm form = new LoginForm("user@example.com", "12345678");
        Set<ConstraintViolation<LoginForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }
}
