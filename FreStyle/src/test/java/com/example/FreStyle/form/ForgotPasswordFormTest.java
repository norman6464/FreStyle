package com.example.FreStyle.form;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class ForgotPasswordFormTest {

    private static Validator validator;

    @BeforeAll
    static void setUp() {
        validator = Validation.buildDefaultValidatorFactory().getValidator();
    }

    @Test
    void 正常なフォームはバリデーションエラーなし() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "123456", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void メールアドレスが空の場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("", "123456", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスを入力してください"));
    }

    @Test
    void メールアドレスの形式が不正な場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("bad-email", "123456", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスは正しい形式で入力してください。"));
    }

    @Test
    void コードが空の場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("コードを入力してください"));
    }

    @Test
    void コードが6文字未満の場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "12345", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("正しい桁数を入力してください"));
    }

    @Test
    void コードが8文字超の場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "123456789", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("正しい桁数を入力してください"));
    }

    @Test
    void 新パスワードが空の場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "123456", "");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("パスワードを入力してください"));
    }

    @Test
    void 新パスワードが8文字未満の場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "123456", "short");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("正しい桁数を入力してください"));
    }

    @Test
    void 新パスワードがちょうど8文字の場合バリデーション通過() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "123456", "12345678");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void メールアドレスがnullの場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm(null, "123456", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスを入力してください"));
    }

    @Test
    void コードがnullの場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", null, "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("コードを入力してください"));
    }

    @Test
    void 新パスワードがnullの場合エラー() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "123456", null);
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("パスワードを入力してください"));
    }

    @Test
    void コードがちょうど6文字の場合バリデーション通過() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "123456", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void コードがちょうど8文字の場合バリデーション通過() {
        ForgotPasswordForm form = new ForgotPasswordForm("user@example.com", "12345678", "newpass123");
        Set<ConstraintViolation<ForgotPasswordForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }
}
