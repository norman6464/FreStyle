package com.example.FreStyle.form;


import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ProfileForm {
  private String username;
  private String email;
  private String bio;
}
