# Growatt API Golang client

[![Go Report Card](https://goreportcard.com/badge/github.com/skoef/growatt)](https://goreportcard.com/report/github.com/skoef/growatt) [![Documentation](https://godoc.org/github.com/skoef/growatt?status.svg)](http://godoc.org/github.com/skoef/growatt) [![Building](https://travis-ci.com/skoef/growatt.svg?branch=master)](https://travis-ci.com/skoef/growatt/)

This is a simple golang library using the rather quirky Growatt API on https://server.growatt.com. This libary is inspired by indykoning's [PyPi_GrowattServer](https://github.com/indykoning/PyPi_GrowattServer) and Sjord's [growatt_api_client](https://github.com/Sjord/growatt_api_client). It tries to normalize objects as much as possible, so API output is parsed and then converted into defined types within the library.

For simplicity sake, currently only several API endpoints are supported. If you miss specific features in the library, please open an issue!

## Example usage
An example for using the API client is shown below, where the credentials are those you would login to https://server.growatt.com/ with:

```golang
package main

import (
  "fmt"

  "github.com/skoef/growatt"
)

func main() {
  api := growatt.NewAPI("johndoe", "s3cr3t")
  plants, err := api.GetPlantList()
  if err != nil {
    panic(err)
  }

  fmt.Printf("found %d plants in your Growatt account\n", len(plants))

  for _, plant := range plants {
    inverters, err := api.GetPlantInverterList(plant.ID)
    if err != nil {
      continue
    }

    fmt.Printf("plant %d has %d inverters\n", plant.ID, len(inverters))
    for _, inverter := range inverters {
      fmt.Printf("inverter %s is generating %0.2fW\n", inverter.Serial, inverter.CurrentPower)
    }
    fmt.Printf("the plant's total combined power is %0.2fW\n", plant.CurrentPower)
  }
}
```
