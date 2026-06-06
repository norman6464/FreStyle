package com.normanblog.frestyle;

import static org.assertj.core.api.Assertions.assertThat;

import org.flywaydb.core.Flyway;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

/** Flyway のマイグレーションが起動時に適用されることを検証する。 */
@SpringBootTest
class MigrationTest {

  @Autowired private Flyway flyway;

  @Test
  void flyway_appliesAllMigrations() {
    // 起動時に全 migration が順に適用され、最新が末尾の版になっていること。
    var applied =
        java.util.Arrays.stream(flyway.info().applied())
            .map(m -> m.getVersion().getVersion())
            .toList();
    assertThat(applied).contains("1", "2", "3", "4", "5");
    assertThat(flyway.info().current().getVersion().getVersion()).isEqualTo("5");
  }
}
