package model

import (
	"strings"
	"testing"
	"time"
)

func TestMarshalPreSecKillRecord(t *testing.T) {
	record := &PreSecKillRecord{
		SecNum:     "sec-1",
		UserID:     10,
		GoodsID:    20,
		OrderNum:   "order-1",
		Price:      18.8,
		Status:     1,
		CreateTime: time.Unix(1700000000, 0),
		ModifyTime: time.Unix(1700000100, 0),
	}

	encoded, err := marshalPreSecKillRecord(record)
	if err != nil {
		t.Fatalf("marshal record failed: %v", err)
	}

	if !strings.Contains(encoded, `"sec-1"`) {
		t.Fatalf("unexpected encoded record: %s", encoded)
	}
}

func TestBuildPreDescStockKeys(t *testing.T) {
	keys := buildPreDescStockKeys(10, 20, 2, "sec-1", `{"ok":true}`)
	if len(keys) != 5 {
		t.Fatalf("expected 5 keys, got %d", len(keys))
	}
	if keys[0] != "10" || keys[1] != "20" || keys[2] != "2" || keys[3] != "sec-1" {
		t.Fatalf("unexpected keys: %#v", keys)
	}
}

func TestBuildSetSuccessKeys(t *testing.T) {
	keys := buildSetSuccessKeys(10, 20, "sec-1", `{"ok":true}`)
	if len(keys) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(keys))
	}
	if keys[0] != "10" || keys[1] != "20" || keys[2] != "sec-1" {
		t.Fatalf("unexpected keys: %#v", keys)
	}
}
