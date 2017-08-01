package torpedo_common

import "fmt"

func FormatTDiff(ts int64) (int64, int64, int64, int64) {
	m, s := ts/60, ts%60
	h, m := m/60, m%60
	d, h := h/24, h%24
	return d, h, m, s
}

func CalculateMessageRate(tdiff, msgcount int64) (result string) {
	var value int64
	pairs := make(map[string]int64)
	pairs["s"] = 0
	pairs["m"] = 60
	pairs["h"] = 3600
	pairs["d"] = 86400
	if msgcount == 0 {
		result = "0/s"
		return
	}
	// BUGGY!
	for key := range pairs {
		if pairs[key] == 0 {
			value = msgcount / tdiff
		} else {
			value = msgcount / (tdiff / pairs[key])
		}
		if value > 0 {
			result = fmt.Sprintf("%v/%s", value, key)
			break
		}
	}
	return
}
