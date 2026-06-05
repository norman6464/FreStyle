package com.normanblog.frestyle.handler;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** NoteController と HealthController の HTTP 振る舞いを検証する。 */
@SpringBootTest
class NoteControllerTest {

  @Autowired private WebApplicationContext context;

  private MockMvc mockMvc() {
    return MockMvcBuilders.webAppContextSetup(context).build();
  }

  @Test
  void health_returnsUp() throws Exception {
    mockMvc()
        .perform(get("/api/v2/health"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.status").value("UP"));
  }

  @Test
  void createThenList_returnsCreatedNote() throws Exception {
    MockMvc mvc = mockMvc();
    mvc.perform(
            post("/api/v2/notes")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"テストノート\",\"content\":\"本文\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.title").value("テストノート"))
        .andExpect(jsonPath("$.isPublic").value(false))
        .andExpect(jsonPath("$.isPinned").value(false))
        // Lombok getter 由来の重複キー(public/pinned)が出ないこと
        .andExpect(jsonPath("$.public").doesNotExist())
        .andExpect(jsonPath("$.pinned").doesNotExist());

    mvc.perform(get("/api/v2/notes"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$[0].title").value("テストノート"));
  }
}
