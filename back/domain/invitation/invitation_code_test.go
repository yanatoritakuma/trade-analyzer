package invitation_test

import (
	"testing"
	"time"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/invitation"
)

func TestInvitationCode_IsValid(t *testing.T) {
	uid := uint(1)
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	cases := []struct {
		name string
		code invitation.InvitationCode
		want bool
	}{
		{"有効", invitation.InvitationCode{IsActive: true, ExpiresAt: future}, true},
		{"無効化", invitation.InvitationCode{IsActive: false, ExpiresAt: future}, false},
		{"使用済", invitation.InvitationCode{IsActive: true, ExpiresAt: future, UsedBy: &uid}, false},
		{"期限切れ", invitation.InvitationCode{IsActive: true, ExpiresAt: past}, false},
	}
	for _, c := range cases {
		if got := c.code.IsValid(); got != c.want {
			t.Errorf("%s: IsValid() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestInvitationCode_MarkAsUsed(t *testing.T) {
	c := invitation.InvitationCode{IsActive: true, ExpiresAt: time.Now().Add(time.Hour)}
	c.MarkAsUsed(42)
	if c.UsedBy == nil || *c.UsedBy != 42 {
		t.Fatal("UsedBy が設定されていない")
	}
	if c.UsedAt == nil {
		t.Fatal("UsedAt が設定されていない")
	}
	if c.IsValid() {
		t.Fatal("使用済みは IsValid() = false であるべき")
	}
}

func TestInvitationCode_Status(t *testing.T) {
	uid := uint(1)
	if (&invitation.InvitationCode{IsActive: false, ExpiresAt: time.Now().Add(time.Hour)}).Status() != "disabled" {
		t.Error("disabled 判定が誤り")
	}
	if (&invitation.InvitationCode{IsActive: true, ExpiresAt: time.Now().Add(time.Hour), UsedBy: &uid}).Status() != "used" {
		t.Error("used 判定が誤り")
	}
	if (&invitation.InvitationCode{IsActive: true, ExpiresAt: time.Now().Add(-time.Hour)}).Status() != "expired" {
		t.Error("expired 判定が誤り")
	}
	if (&invitation.InvitationCode{IsActive: true, ExpiresAt: time.Now().Add(time.Hour)}).Status() != "valid" {
		t.Error("valid 判定が誤り")
	}
}
