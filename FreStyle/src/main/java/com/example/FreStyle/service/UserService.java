package com.example.FreStyle.service;

import java.util.Optional;

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
    user.setIsActive(false);
    userRepository.save(user);
  }
  
  @Transactional
  public void registerUserOIDC(String name, String email, String sub) {
    if (userRepository.existsByEmail(email)) {
      System.out.println("already exist email");
      return;
    }
    
    User user = new User();
    user.setUsername(name);
    user.setEmail(email);
    user.setCognitoSub(sub);
    userRepository.save(user);
    
  }
  

  // ユーザーを有効にする
  @Transactional
  public void activeUser(String email) {
    User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("ユーザーが見つかりません。"));

    if (Boolean.TRUE.equals(user.getIsActive())) {
      System.out.println("メール認証は完了していないため、ログインできません。");
      return;
    }

    user.setIsActive(true);
    userRepository.save(user);
  }

  @Transactional(readOnly = true)
  public void checkUserIsActive(String email) {
    User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("ユーザーは存在しません。"));
    if (!Boolean.TRUE.equals(user.getIsActive())) {
      throw new RuntimeException("メール認証は完了していないため、ログインできません。");
    }
  }

  @Transactional
  public void registerCognitoSubject(String sub, String email) {
    User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("ユーザーは存在しません。"));

    if (!Boolean.TRUE.equals(user.getIsActive())) {
      throw new RuntimeException("メール認証は完了していないため、ログインできません。");
    }
    
    user.setCognitoSub(sub);
    userRepository.save(user);
  }
  
  // subからidを探す
  public Integer findUserIdByCognitoSub(String sub) {
     
    return userRepository.findIdByCognitoSub(sub)
    .orElseThrow(() -> new RuntimeException("このリクエストは無効です。"));
    
  }

}
