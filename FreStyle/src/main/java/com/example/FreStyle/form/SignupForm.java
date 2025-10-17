package com.example.FreStyle.form;

import jakarta.validation.constraints.NotBlank;

public class SignupForm {

  @NotBlank(message = "メールアドレスを入力してください")
  private String email;

  @NotBlank(message = "パスワードを入力してください")
  private String password;

  @NotBlank(message = "ユーザー名を入力してください")
  private String name;

  // getter
  public String getEmail() {
    return email;
  }

  // setter
  public void setEmail(String email) {
    this.email = email;
  }

  // getter
  public String getPassword() {
    return password;
  }

  // setter
  public void setPassword(String password) {
    this.password = password;
  }

  // getter
  public String getName() {
    return name;
  }

  // setter
  public void setName(String name) {
    this.name = name;
  }
}
