package util

import (
	"testing"
	"time"

	"cylawcase/internal/constants"
)

func TestGenerateTokenAndParse(t *testing.T) {
	secret := "test-secret-123456"
	token, err := GenerateToken(secret, time.Hour, 7, "lawyer", constants.RoleLawyer)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.UserID != 7 || claims.Username != "lawyer" || claims.Role != constants.RoleLawyer {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	for _, tc := range []string{"", "bad", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig"} {
		if _, err := ParseToken("secret", tc); err == nil {
			t.Errorf("expected error for %q", tc)
		}
	}
}

func TestFormatters(t *testing.T) {
	if got := CaseStatusText(constants.CaseStatusClosed); got != "已结案" {
		t.Errorf("CaseStatusText = %q", got)
	}
	if got := BillingStatusText(constants.BillingStatusPaid); got != "已支付" {
		t.Errorf("BillingStatusText = %q", got)
	}
	if got := BillingTypeText(constants.BillingTypeAttorneyFee); got != "律师费" {
		t.Errorf("BillingTypeText = %q", got)
	}
	if got := CaseTypeText(constants.CaseTypeCivil); got != "民事" {
		t.Errorf("CaseTypeText = %q", got)
	}
}

func TestFormatAmount(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0.00"}, {1234.5, "1,234.50"}, {1234567.89, "1,234,567.89"}, {50000, "50,000.00"},
	}
	for _, tc := range cases {
		if got := FormatAmount(tc.in); got != tc.want {
			t.Errorf("FormatAmount(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
