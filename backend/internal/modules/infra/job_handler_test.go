package infra

import (
	"testing"
	"time"
)

func TestCronMatchesQuartzAndStandardExpressions(t *testing.T) {
	at := time.Date(2026, time.July, 22, 19, 10, 0, 0, time.Local)
	for _, expression := range []string{"*/5 * * * *", "0 */5 * * * ?", "0 */5 * * * ? 2026"} {
		if !cronMatches(expression, at) {
			t.Fatalf("expected %q to match %s", expression, at)
		}
	}
	if cronMatches("0 */7 * * * ?", at) {
		t.Fatal("unexpected seven-minute expression match")
	}
}

func TestNextCronTimes(t *testing.T) {
	from := time.Date(2026, time.July, 22, 19, 8, 12, 0, time.Local)
	times, err := nextCronTimes("0 */5 * * * ?", from, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []time.Time{
		time.Date(2026, time.July, 22, 19, 10, 0, 0, time.Local),
		time.Date(2026, time.July, 22, 19, 15, 0, 0, time.Local),
		time.Date(2026, time.July, 22, 19, 20, 0, 0, time.Local),
	}
	for index := range want {
		if !times[index].Equal(want[index]) {
			t.Fatalf("next time %d = %s, want %s", index, times[index], want[index])
		}
	}
}
