package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import org.springframework.stereotype.Service;

/** 認証に付随するユーザー状態の同期を担うサービス。 */
@Service
public class AuthService {

  private final UserRepository users;

  public AuthService(UserRepository users) {
    this.users = users;
  }

  /**
   * Cognito の admin グループに属するのに DB の role が管理者でない場合、super_admin へ昇格させる。
   *
   * <p>Cognito 側でグループ付与した管理者を、毎回コンソールで role 設定し直さずに済ませるための
   * 自動同期(既存 Go 実装の /auth/me と同じ挙動)。昇格が不要なら何もしない。
   */
  public User syncAdminRole(User user, boolean isAdmin) {
    boolean alreadyAdmin =
        Role.SUPER_ADMIN.equals(user.getRole()) || Role.COMPANY_ADMIN.equals(user.getRole());
    if (isAdmin && !alreadyAdmin) {
      user.setRole(Role.SUPER_ADMIN);
      user.setUpdatedAt(Instant.now());
      return users.save(user);
    }
    return user;
  }
}
