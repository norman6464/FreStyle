package com.example.FreStyle.form;


import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ProfileForm {
  @NotBlank(message = "ユーザー名を入力してください")
  private String name;
  private String bio;
  private String iconUrl;
}
