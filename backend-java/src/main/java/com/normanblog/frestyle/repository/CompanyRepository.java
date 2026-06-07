package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.Company;
import org.springframework.data.jpa.repository.JpaRepository;

/** companies テーブルへのアクセスを担うリポジトリ。 */
public interface CompanyRepository extends JpaRepository<Company, Long> {}
