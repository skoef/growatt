package growatt

// Inverter represents the datastructure for a Growatt inverter
type Inverter struct {
	IsLost        bool
	InverterType  string
	EnergyToday   float64 // in kilowatthours
	EnergyTotal   float64 // in kilowatthours
	Location      string
	Alias         string
	Type          string
	DatalogSerial string
	Serial        string
	CurrentPower  float64 // in watts
	Status        int
}

type inverterData struct {
	Lost          bool   `json:"lost"`
	InverterType  string `json:"invType"`
	EnergyToday   string `json:"eToday"`
	Location      string `json:"location"`
	Alias         string `json:"deviceAilas"`
	Type          string `json:"deviceType"`
	DatalogSerial string `json:"datalogSn"`
	Serial        string `json:"deviceSn"`
	CurrentPower  string `json:"power"`
	Status        int    `json:"deviceStatus"`
	EnergyTotal   string `json:"energy"`
}

func parseInvertedData(i inverterData) Inverter {
	return Inverter{
		IsLost:        i.Lost,
		InverterType:  i.InverterType,
		EnergyToday:   parsePower(i.EnergyToday),
		EnergyTotal:   parsePower(i.EnergyTotal),
		Location:      i.Location,
		Alias:         i.Alias,
		Type:          i.Type,
		DatalogSerial: i.DatalogSerial,
		Serial:        i.Serial,
		CurrentPower:  parsePower(i.CurrentPower),
		Status:        i.Status,
	}
}
