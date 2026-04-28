package handler

import (
	"testing"
	"time"
)

func TestBuildChatMessage(t *testing.T) {
	now := time.Date(2026, 4, 28, 6, 30, 0, 0, time.UTC)
	got, ok := BuildChatMessage(ChatInbound{Type: "send", Content: "hi"}, "42", 7, "alice", now)
	if !ok {
		t.Fatal("should succeed for valid send")
	}
	if got.Type != "message" {
		t.Errorf("Type = %q", got.Type)
	}
	if got.SenderID != 7 || got.SenderName != "alice" {
		t.Errorf("sender mismatch: %+v", got)
	}
	if got.RoomID != "42" {
		t.Errorf("RoomID = %q", got.RoomID)
	}
	if got.Content != "hi" {
		t.Errorf("Content = %q", got.Content)
	}
	if got.CreatedAt == "" {
		t.Error("CreatedAt should be set")
	}
}

func TestBuildChatMessage_RejectsNonSend(t *testing.T) {
	if _, ok := BuildChatMessage(ChatInbound{Type: "delete", Content: "x"}, "1", 1, "a", time.Now()); ok {
		t.Fatal("should reject non-send type")
	}
}

func TestBuildChatMessage_RejectsEmptyContent(t *testing.T) {
	if _, ok := BuildChatMessage(ChatInbound{Type: "send"}, "1", 1, "a", time.Now()); ok {
		t.Fatal("should reject empty content")
	}
}

func TestBuildChatDelete(t *testing.T) {
	got, ok := BuildChatDelete(ChatInbound{Type: "delete", CreatedAtRef: "2026-04-28T00:00:00Z"}, "42")
	if !ok {
		t.Fatal("should succeed for valid delete")
	}
	if got.Type != "delete" || got.RoomID != "42" || got.CreatedAt != "2026-04-28T00:00:00Z" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestBuildChatDelete_RequiresRef(t *testing.T) {
	if _, ok := BuildChatDelete(ChatInbound{Type: "delete"}, "1"); ok {
		t.Fatal("should reject empty CreatedAtRef")
	}
}

func TestDecodeInbound(t *testing.T) {
	in, err := DecodeInbound([]byte(`{"type":"send","content":"hello"}`))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if in.Type != "send" || in.Content != "hello" {
		t.Errorf("decoded: %+v", in)
	}
}

func TestEncodeOutbound(t *testing.T) {
	raw, err := EncodeOutbound(ChatOutbound{Type: "message", Content: "hi"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(raw) == "" {
		t.Fatal("empty json")
	}
}
