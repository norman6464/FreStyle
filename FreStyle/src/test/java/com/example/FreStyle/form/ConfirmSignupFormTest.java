package com.example.FreStyle.form;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class ConfirmSignupFormTest {

    private static Validator validator;

    @BeforeAll
    static void setUp() {
        validator = Validation.buildDefaultValidatorFactory().getValidator();
    }

    @Test
    void 正常なフォームはバリデーションエラーなし() {
        ConfirmSignupForm form = new ConfirmSignupForm("user@example.com", "123456");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void メールアドレスが空の場合エラー() {
        ConfirmSignupForm form = new ConfirmSignupForm("", "123456");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスを入力してください"));
    }

    @Test
    void メールアドレスの形式が不正な場合エラー() {
        ConfirmSignupForm form = new ConfirmSignupForm("invalid", "123456");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("メールアドレスは正しい形式で入力してください。"));
    }

    @Test
    void コードが空の場合エラー() {
        ConfirmSignupForm form = new ConfirmSignupForm("user@example.com", "");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("コードを入力してください"));
    }

    @Test
    void コードが6文字未満の場合エラー() {
        ConfirmSignupForm form = new ConfirmSignupForm("user@example.com", "12345");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("正しい桁数を入力してください"));
    }

    @Test
    void コードが8文字超の場合エラー() {
        ConfirmSignupForm form = new ConfirmSignupForm("user@example.com", "123456789");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).anyMatch(v -> v.getMessage().equals("正しい桁数を入力してください"));
    }

    @Test
    void コードがちょうど6文字の場合バリデーション通過() {
        ConfirmSignupForm form = new ConfirmSignupForm("user@example.com", "123456");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }

    @Test
    void コードがちょうど8文字の場合バリデーション通過() {
        ConfirmSignupForm form = new ConfirmSignupForm("user@example.com", "12345678");
        Set<ConstraintViolation<ConfirmSignupForm>> violations = validator.validate(form);
        assertThat(violations).isEmpty();
    }
}
