package com.example.FreStyle.service;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.LoginUserDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.*;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.UserRepository;

@Service
public class UserService {
  private final UserRepository userRepository;
  private final ChatRoomRepository chatRoomRepository;

  public UserService(UserRepository userRepository, ChatRoomRepository chatRoomRepository) {
    this.userRepository = userRepository;
    this.chatRoomRepository = chatRoomRepository;
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

  // OIDCログインでDBにユーザ登録をする
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

  // ユーザーが有効化を確認する
  @Transactional(readOnly = true)
  public void checkUserIsActive(String email) {
    User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("ユーザーは存在しません。"));
    if (!Boolean.TRUE.equals(user.getIsActive())) {
      throw new RuntimeException("メール認証は完了していないため、ログインできません。");
    }
  }

  // ログイン時にJWTトークンのsub値を登録をする
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

  public User findUser(String cognitoSub) {
    return userRepository.findByCognitoSub(cognitoSub)
        .orElseThrow(() -> new RuntimeException("User not found"));
  }

  // subからidを探す
  public Integer findUserIdByCognitoSub(String sub) {

    return userRepository.findIdByCognitoSub(sub)
        .orElseThrow(() -> new RuntimeException("このリクエストは無効です。"));

  }
  
  // subからログイン済かどうかを返す
  public LoginUserDto findLoginUserByCognitoSub(String sub) {
    User user = userRepository.findByCognitoSub(sub)
    .orElseThrow(() -> new RuntimeException("ユーザーが存在しません。"));
    LoginUserDto dto = new LoginUserDto();
    dto.setSub(user.getCognitoSub());
    dto.setName(user.getUsername());
    dto.setEmail(user.getEmail());
    return dto;
  }

  // 部分一致でユーザーのemailを検索をする
  public List<UserDto> findUsersWithRoomId(Integer id, String email) {
    List<UserDto> users;

    if (email == null || email.isEmpty()) {
      users = userRepository.findAllUserDtos(id);
    } else {
      String queryEmail = "%" + email + "%";
      users = userRepository.findIdAndEmailByEmailLikeDtos(id, queryEmail);
    }

    for (UserDto user : users) {
      Integer roomId = chatRoomRepository.findRoomIdByUserIds(id, user.getId());
      user.setRoomId(roomId); // まだルームがない場合は null
    }

    return users;
  }


  @Transactional
  public void updateUser(ProfileForm form, String sub) {
    User user = userRepository.findByCognitoSub(sub)
        .orElseThrow(() -> new RuntimeException("ユーザーが見つかりません。"));
        
    user.setUsername(form.getUsername());
    user.setBio(form.getBio());
    
    userRepository.save(user);
  }


}
