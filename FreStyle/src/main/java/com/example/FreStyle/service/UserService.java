package com.example.FreStyle.service;


import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.*;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.repository.UserRepository;

@Service
public class UserService {
    private final UserRepository userRepository;
    
    public UserService(UserRepository userRepository) {
      this.userRepository = userRepository;
    }
    
    // ユーザー情報登録
    @Transactional
    public void registerUser(SignupForm form) {
      if (userRepository.existsByEmail(form.getEmail())) {
        throw new RuntimeException("このメールアドレスは既に登録されています。");
      }
      
      if (userRepository.existsByUsername(form.getName())) {
        throw new RuntimeException("このユーザー名は既に使用されています。");
      }
      
      User user = new User();
      user.setUsername(form.getName());
      user.setEmail(form.getEmail());
      // Bcryptで暗号化したパスワードを保存する
      // user.setPasswordHash(PasswordUtils.hash(form.getPassword()));
      // ほかのフィールドはデフォルト値のままでOKになる
      
      userRepository.save(user);
    }
    
    // ユーザーを有効にする
    @Transactional
    public void activeUserByEmail(String email) {
      User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("指定されたメールアドレスのユーザーが見つかりません。"));
        
      if (Boolean.TRUE.equals(user.getIsActive())) {
        System.out.println("this email  is verified");
        return;
      }
      
      user.setIsActive(true);
      userRepository.save(user);
    }
    
    @Transactional(readOnly = true)
    public void checkUserIsActiveByEmail(String email) {
      User user = userRepository.findByEmail(email)
          .orElseThrow(() -> new RuntimeException("ユーザーは存在しません。"));
      if (!Boolean.TRUE.equals(user.getIsActive())) {
        throw new RuntimeException("メール認証は完了していないため、ログインできません。");
      }
    }
  
}
