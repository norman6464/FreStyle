package com.example.FreStyle.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.UserIdentity;

public interface UserIdentityRepository extends JpaRepository<UserIdentity, Integer> {
  
   // provider + sub でUserIdentityを取得
    Optional<UserIdentity> findByProviderAndProviderSub(String provider, String providerSub);
    
    // subで検索をする
    Optional<UserIdentity> findByProviderSub(String providerSub);

}
