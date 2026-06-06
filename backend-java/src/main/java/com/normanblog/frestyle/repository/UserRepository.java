package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.User;
import java.util.Optional;
import org.springframework.data.jpa.repository.JpaRepository;

/** users テーブルへのアクセスを担うリポジトリ。 */
public interface UserRepository extends JpaRepository<User, Long> {

  // 論理削除されていない行のみを認証対象とする。
  Optional<User> findByCognitoSubAndDeletedAtIsNull(String cognitoSub);
}
