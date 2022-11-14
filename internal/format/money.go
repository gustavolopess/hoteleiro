package format

import (
	"strconv"
	"strings"
)

func BrlToFloat64(brlValue string) (float64, error) {
	brlValue = strings.TrimPrefix(brlValue, "R$")
	brlValue = strings.ReplaceAll(brlValue, ".", "")
	brlValue = strings.ReplaceAll(brlValue, ",", ".")
	converted, err := strconv.ParseFloat(brlValue, 64)
	if err != nil {
		return -1, err
	}
	return converted, err
}
