package qnaccount

import (
	"reflect"
	"testing"
)

func TestHasEndpoint(t *testing.T) {
	a := &Account{
		EndpointIds: []string{
			"a1", "b2", "c3",
		},
	}

	if !a.HasEndpointId("b2") {
		t.Fatal("expected true")
	}

	if a.HasEndpointId("d4") {
		t.Fatal("expected false")
	}
}

func TestSetFromAccountData(t *testing.T) {
	ad := &AccountData{
		QuicknodeId: "some-id",
		EndpointId:  "some-endpoint-id",
		Plan:        "plan-slug",
		WssUrl:      "wss://example.com",
		HttpUrl:     "http://example.com",
		Chain:       "ethereum",
		Network:     "mainnet",
	}

	a := &Account{}
	a.SetFromAccountData(ad)

	if v := a.QuicknodeId; v != ad.QuicknodeId {
		t.Fatal("wrong QuicknodeId", v)
	}
	if v := a.Plan; v != ad.Plan {
		t.Fatal("wrong Plan", v)
	}
	if v := a.WssUrl; v != ad.WssUrl {
		t.Fatal("wrong WssUrl", v)
	}
	if v := a.HttpUrl; v != ad.HttpUrl {
		t.Fatal("wrong HttpUrl", v)
	}
	if v := a.Chain; v != ad.Chain {
		t.Fatal("wrong Chain", v)
	}
	if v := a.Network; v != ad.Network {
		t.Fatal("wrong Network", v)
	}
}

func TestDeactivateEndpoint(t *testing.T) {
	a := &Account{
		EndpointIds: []string{
			"a1", "b2", "c3",
		},
	}

	if found := a.DeactivateEndpoint("z10"); found {
		t.Fatal("reports found = true for missing item")
	}
	if found := a.DeactivateEndpoint("a1"); !found {
		t.Fatal("should report found = true")
	}
	if !reflect.DeepEqual(a.EndpointIds, []string{"b2", "c3"}) {
		t.Fatal("wrong EndpointIds", a.EndpointIds)
	}
}
