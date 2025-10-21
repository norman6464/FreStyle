package com.example.FreStyle.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.User;

@Repository
public interface UserRepository extends JpaRepository<User, Integer>{
  
  boolean existsByEmail(String email);
  
  boolean existsByUsername(String username);
  
  Optional<User> findByEmail(String email);
  
  @Query("SELECT u.id FROM User u WHERE u.cognitoSub = :sub")
  Optional<Integer> findIdByCognitoSub(@Param("sub")String sub);
  
}
