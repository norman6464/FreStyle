package com.example.FreStyle.form;

import org.hibernate.validator.constraints.Length;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class SignupForm {

  @NotBlank(message = "メールアドレスを入力してください")
  @Email(message = "メールアドレスは正しい形式で入力してください。")
  private String email;

  @NotBlank(message = "パスワードを入力してください")
  @Length(min = 8, message = "正しい桁数を入力してください")
  private String password;

  @NotBlank(message = "ユーザー名を入力してください")
  private String name;
}
