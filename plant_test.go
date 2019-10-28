package growatt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlantData(t *testing.T) {
	p := parsePlantData(plantData{
		PlantMoneyText: "foo",
		PlantName:      "bar",
		PlantID:        "1234",
		IsHaveStorage:  "true",
		TodayEnergy:    "123.4 kWh",
		TotalEnergy:    "456.7 kWH",
		CurrentPower:   "890.1 W",
	})

	assert.Equal(t, "foo", p.Earned)
	assert.Equal(t, "bar", p.Name)
	assert.Equal(t, 1234, p.ID)
	assert.Equal(t, true, p.HasStorage)
	assert.Equal(t, 123.4, p.EnergyToday)
	assert.Equal(t, 456.7, p.EnergyTotal)
	assert.Equal(t, 890.1, p.CurrentPower)
}
