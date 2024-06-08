package main

import (
	"burlo/pkg/ntfy"
	"fmt"
)

var notify ntfy.Notify

func initNotifyClient(ntfyHost string) {
	notify = ntfy.New(ntfyHost, "burlo")
}

func notifyMode(mode dx2wmode) {
	if mode == DX2W_HEAT {
		notify.Publish(
			"Heating mode activated",
			"Its getting chilly out there",
			[]string{"house_with_garden", "fire"},
		)
	} else {
		notify.Publish(
			"Cooling mode activated",
			"Wow its hot out there",
			[]string{"house_with_garden", "snowflake"},
		)
	}
}

func notifyState(state dx2wstate) {
	if state == DX2W_OFF {
		notify.Publish(
			"DX2W Standby",
			"Saves energy when there is no need to heat or cool for long periods of time. "+
				"Buffer temperature will not be maintained while in standby.",
			[]string{"house_with_garden", "zzz"},
		)
	}
}

func notifyWindow(window wmode) {
	if window == OPEN {
		notify.Publish(
			"Its nice out there!",
			fmt.Sprintf("Now is a good time to open those windows and get some fresh air. %.1f°C, %.0f%% relH, AQHI: %d",
				inputs.Outdoor.Temperature,
				inputs.Outdoor.Humidity,
				inputs.Outdoor.AQHI),
			[]string{"house_with_garden", "sun_behind_small_cloud"},
		)
	} else {
		notify.Publish(
			"Keep windows closed",
			fmt.Sprintf("%.1f°C, %.0f%% relH, AQHI: %d",
				inputs.Outdoor.Temperature,
				inputs.Outdoor.Humidity,
				inputs.Outdoor.AQHI),
			[]string{"house_with_garden", "window"},
		)
	}
}
