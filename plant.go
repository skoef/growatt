package growatt

import (
	"regexp"
	"strconv"
	"time"
)

// Plant represents the data structure for a Growatt plant
type Plant struct {
	Earned       string
	Name         string
	ID           int
	HasStorage   bool
	EnergyToday  float64 // in kilowatthours
	EnergyTotal  float64 // in kilowatthours
	CurrentPower float64 // in watts
}

// PlantEnergy is the amount of power the plant generated at a certain time
type PlantEnergy struct {
	Timestamp time.Time
	Power     float64 // in watts
}

// plantData represents how plant data is returned from the API
type plantData struct {
	PlantMoneyText string `json:"plantMoneyText"`
	PlantName      string `json:"plantName"`
	PlantID        string `json:"plantId"`
	IsHaveStorage  string `json:"isHaveStorage"`
	TodayEnergy    string `json:"todayEnergy"`
	TotalEnergy    string `json:"totalEnergy"`
	CurrentPower   string `json:"currentPower"`
}

func parsePower(pwr string) float64 {
	wattRe := regexp.MustCompile(`^(?i)([0-9.]+)(?: k?Wh?)?$`)
	match := wattRe.FindStringSubmatch(pwr)
	if len(match) != 2 {
		return 0.0
	}

	result, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0.0
	}

	return result
}

func parsePlantData(p plantData) Plant {
	plantID, _ := strconv.ParseInt(p.PlantID, 10, 64)

	return Plant{
		Earned:       p.PlantMoneyText,
		Name:         p.PlantName,
		ID:           int(plantID),
		HasStorage:   (p.IsHaveStorage == "true"),
		EnergyToday:  parsePower(p.TodayEnergy),
		EnergyTotal:  parsePower(p.TotalEnergy),
		CurrentPower: parsePower(p.CurrentPower),
	}
}
