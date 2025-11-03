package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.User;

@Repository
public interface UserRepository extends JpaRepository<User, Integer> {

  boolean existsByEmail(String email);

  boolean existsByUsername(String username);

  Optional<User> findByEmail(String email);
  
  Optional<User> findByCognitoSub(String cognitoSub);

  @Query("SELECT u.id FROM User u WHERE u.cognitoSub = :sub")
  Optional<Integer> findIdByCognitoSub(@Param("sub") String sub);
  
  // 部分一致検索
  @Query("SELECT new com.example.FreStyle.dto.UserDto(u.id, u.email) FROM User u WHERE u.email LIKE :email AND u.id <> :id")
  List<UserDto> findIdAndEmailByEmailLikeDtos(@Param("id") Integer id,@Param("email") String email);

  // 全件取得
  @Query("SELECT new com.example.FreStyle.dto.UserDto(u.id, u.email) FROM User u WHERE u.id <> :id")
  List<UserDto> findAllUserDtos(@Param("id") Integer id);

}
