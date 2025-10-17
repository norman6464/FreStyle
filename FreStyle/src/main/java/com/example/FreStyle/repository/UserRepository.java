package com.example.FreStyle.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.User;

@Repository
public interface UserRepository extends JpaRepository<User, Integer>{
  boolean existsByEmail(String email);
  boolean existsByUsername(String username);
  Optional<User> findByEmail(String email);
}
