package com.example.FreStyle.form;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Data;

// DTO
@AllArgsConstructor
@Data
public class SignupForm {
  
  @NotBlank(message = "メールアドレスを入力してください")
  private String email;
  
  @NotBlank(message = "パスワードを入力してください")
  private String password;
  
  @NotBlank(message = "ユーザー名を入力してください")
  private String name;

}
