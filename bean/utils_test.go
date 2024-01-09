package bean

import (
	"testing"
	"time"
)

func Test_getDate(t *testing.T) {
	res, _ := getDate("2022-01-01")
	expected := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	if res != expected {
		t.Errorf("incorrect result: expected %s, got %s", expected, res)
	}

  res, err := getDate("invalid")
	if err == nil {
		t.Errorf("incorrect result: expected error")
	}
}
