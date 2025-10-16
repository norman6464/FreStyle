package com.example.FreStyle.form;

import lombok.AllArgsConstructor;
import lombok.Data;

@AllArgsConstructor
@Data
public class ConfirmSignupForm {
  private String email;
  private String code;
}
