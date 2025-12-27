package com.example.FreStyle.service;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.LoginUserDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.UserRepository;

@Service
public class UserService {

  private final UserRepository userRepository;
  private final UserIdentityService userIdentityService;
  private final ChatRoomRepository chatRoomRepository;

  public UserService(UserRepository userRepository,
      UserIdentityService userIdentityService,
      ChatRoomRepository chatRoomRepository) {

    this.userRepository = userRepository;
    this.userIdentityService = userIdentityService;
    this.chatRoomRepository = chatRoomRepository;
  }

  // ------------------------
  // 通常登録
  // ------------------------
  @Transactional
  public void registerUser(SignupForm form) {
    if (userRepository.existsByEmail(form.getEmail())) {
      throw new RuntimeException("このメールアドレスは既に使用されています。");
    }

    User user = new User();
    user.setName(form.getName());
    user.setEmail(form.getEmail());
    user.setIsActive(false);
    userRepository.save(user);
  }

  // ------------------------
  // Email で検索
  // ------------------------
  public User findUserByEmail(String email) {
    User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("ユーザーが存在しません。"));

    if (!Boolean.TRUE.equals(user.getIsActive())) {
      throw new RuntimeException("メール認証は完了していないため、ログインできません。");
    }

    return user;
  }

  // ------------------------
  // OIDC ログイン時、User を作成し、Identity を追加
  // ------------------------
  @Transactional
  public void registerUserOIDC(String name, String email, String provider, String sub) {

    User user = userRepository.findByEmail(email).orElse(null);

    if (user == null) {
      user = new User();
      user.setName(name);
      user.setEmail(email);
      user.setIsActive(true);
      userRepository.save(user);
    }

    // UserIdentityService に責務を委譲
    userIdentityService.registerUserIdentity(user, provider, sub);
  }

  // ------------------------
  // sub から UserDto を返す
  // ------------------------
  public LoginUserDto findLoginUserBySub(String sub) {

    User user = userIdentityService.findUserBySub(sub);

    LoginUserDto dto = new LoginUserDto();
    dto.setSub(sub);
    dto.setName(user.getName());
    dto.setEmail(user.getEmail());

    return dto;
  }

  // ------------------------
  // Profile 更新
  // ------------------------
  @Transactional
  public void updateUser(ProfileForm form, String sub) {

    User user = userIdentityService.findUserBySub(sub);
    user.setName(form.getName());
    user.setBio(form.getBio());

    userRepository.save(user);
  }

  // ------------------------
  // 一覧取得 + Room ID
  // ------------------------
  public List<UserDto> findUsersWithRoomId(Integer id, String query) {

    List<UserDto> users;

    if (query == null || query.isEmpty()) {
      users = userRepository.findAllUserDtos(id);
    } else {
      String queryEmail = "%" + query + "%";
      users = userRepository.findIdAndEmailByEmailLikeDtos(id, queryEmail);
    }

    for (UserDto user : users) {
      Integer roomId = chatRoomRepository.findRoomIdByUserIds(id, user.getId());
      user.setRoomId(roomId);
    }

    return users;
  }

  // ------------------------
  // User を active = true にする
  // ------------------------
  @Transactional
  public void activeUser(String email) {
    User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("ユーザーが見つかりません。"));

    if (Boolean.TRUE.equals(user.getIsActive())) {
      return;
    }

    user.setIsActive(true);
    userRepository.save(user);
  }

  // ------------------------
  // active 状態チェック
  // ------------------------
  @Transactional(readOnly = true)
  public void checkUserIsActive(String email) {

    User user = userRepository.findByEmail(email)
        .orElseThrow(() -> new RuntimeException("ユーザーは存在しません。"));

    if (!Boolean.TRUE.equals(user.getIsActive())) {
      throw new RuntimeException("メール認証は完了していないためログインできません。");
    }
  }

  // ユーザー総数を取得
  @Transactional(readOnly = true)
  public Long getTotalUserCount() {
    return userRepository.count();
  }
}
