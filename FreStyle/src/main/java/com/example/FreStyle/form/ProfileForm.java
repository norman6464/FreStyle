package com.example.FreStyle.form;


import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ProfileForm {
  @NotBlank(message = "ユーザー名を入力してください")
  private String name;
  @Size(max = 500, message = "自己紹介は500文字以内で入力してください")
  private String bio;
  @Size(max = 2048, message = "アイコンURLが長すぎます")
  private String iconUrl;
  @Size(max = 100, message = "ステータスは100文字以内で入力してください")
  private String status;
}
