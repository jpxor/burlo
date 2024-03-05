package main

import "time"

func update_tstat_history(tstat Thermostat) {
	datum := HistoryData{
		SensorID:    tstat.ID,
		Temperature: tstat.State.Temperature,
		Humidity:    tstat.State.Humidity,
		DewPoint:    tstat.State.DewPoint,
		SetpointErr: tstat.SetpointErr,
		Time:        tstat.State.Time,
	}
	history, lbk := global.history.Take()

	history = append(history, datum)

	// only keep history for 24 hours
	for i, datum := range history {
		if time.Since(datum.Time) <= 24*time.Hour {
			history = history[i:]
			break
		}
	}
	global.history.Put(history, lbk)
}
