package service

import (
	"testing"

	"cylawcase/internal/constants"
)

func TestCanFlow(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{constants.CaseStatusFiled, constants.CaseStatusInvestigating, true},
		{constants.CaseStatusInvestigating, constants.CaseStatusHearing, true},
		{constants.CaseStatusHearing, constants.CaseStatusClosed, true},
		{constants.CaseStatusClosed, constants.CaseStatusArchived, true},
		{constants.CaseStatusFiled, constants.CaseStatusClosed, false},
		{constants.CaseStatusClosed, constants.CaseStatusFiled, false},
		{constants.CaseStatusInvestigating, constants.CaseStatusFiled, true},
	}
	for _, tc := range cases {
		if got := canFlow(tc.from, tc.to); got != tc.want {
			t.Errorf("canFlow(%s->%s) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestContains(t *testing.T) {
	if !contains(constants.CaseTypeValues, constants.CaseTypeLabor) {
		t.Error("labor should be in case types")
	}
	if contains(constants.CaseTypeValues, "bogus") {
		t.Error("bogus should not be in case types")
	}
}

func TestU64(t *testing.T) {
	if u64(42) != "42" {
		t.Error("u64(42) != 42")
	}
}

func TestStatusValidators(t *testing.T) {
	if !constants.IsValidCaseStatus(constants.CaseStatusArchived) {
		t.Error("archived should be valid")
	}
	if constants.IsValidCaseStatus("bogus") {
		t.Error("bogus should be invalid")
	}
	if !constants.IsValidBillingStatus(constants.BillingStatusInvoiced) {
		t.Error("invoiced should be valid")
	}
	if !constants.IsValidBillingType(constants.BillingTypeTravelFee) {
		t.Error("travel_fee should be valid")
	}
}
