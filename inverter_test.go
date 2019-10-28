package growatt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInverterData(t *testing.T) {
	i := parseInvertedData(inverterData{
		Lost:          true,
		InverterType:  "foo",
		EnergyToday:   "1.2",
		Location:      "Shanghai",
		Alias:         "My first inverter",
		Type:          "inverter",
		DatalogSerial: "1234AB",
		Serial:        "2345CD",
		CurrentPower:  "2.3",
		Status:        6,
		EnergyTotal:   "4.5",
	})

	assert.Equal(t, true, i.IsLost)
	assert.Equal(t, "foo", i.InverterType)
	assert.Equal(t, 1.2, i.EnergyToday)
	assert.Equal(t, "Shanghai", i.Location)
	assert.Equal(t, "My first inverter", i.Alias)
	assert.Equal(t, "inverter", i.Type)
	assert.Equal(t, "1234AB", i.DatalogSerial)
	assert.Equal(t, "2345CD", i.Serial)
	assert.Equal(t, 2.3, i.CurrentPower)
	assert.Equal(t, 6, i.Status)
	assert.Equal(t, 4.5, i.EnergyTotal)
}
