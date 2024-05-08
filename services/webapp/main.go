package main

import (
	"burlo/pkg/dx2w"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type value_formatter_func func(map[string]dx2w.Value) map[string]string

func main() {

	dx2wlogAddr := "192.168.50.193:4006"
	units_temperature := "celcius"
	units_heat := "btu"
	formatter := create_value_formatter_func(units_temperature, units_heat)

	http.HandleFunc("/", indexHandler(dx2wlogAddr, formatter))
	http.HandleFunc("/state-diagram", getStateDiagram(dx2wlogAddr, formatter))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("www/static"))))

	fmt.Println("Server listening on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func getStateDiagram(addr string, formatter value_formatter_func) http.HandlerFunc {
	request_url := fmt.Sprintf("http://%s/dx2w/registers", addr)
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("www/templates/state-diagram.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := http.Get(request_url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var data map[string]dx2w.Value
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, formatter(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func indexHandler(addr string, formatter value_formatter_func) http.HandlerFunc {
	request_url := fmt.Sprintf("http://%s/dx2w/registers", addr)

	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("www/templates/index.html", "www/templates/state-diagram.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := http.Get(request_url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var data map[string]dx2w.Value
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, formatter(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func create_value_formatter_func(units_temp, units_heat string) value_formatter_func {
	return func(valmap map[string]dx2w.Value) map[string]string {
		formatted := make(map[string]string)
		for key, val := range valmap {
			formatted[key] = formatVal(units_temp, units_heat, key, val)
		}

		// rename the key to conform to valid html template names
		formatted["LIQUID_SUB_COOLING"] = formatted["LIQUID_SUB-COOLING"]
		formatted["HP_WATER_DELTA_T"] = formatted["HP_WATER_DELTA-T"]
		formatted["DIVERSION_VALVE_PERCENT_CLOSED"] = formatted["DIVERSION_VALVE_%_CLOSED"]

		// calculate it
		formatted["DISTRIBUTION_FLOW"] = "6 gpm?"

		dt := valmap["MIX_WATER_TEMP"].Float32 - valmap["RETURN_WATER_TEMP"].Float32
		formatted["HEAT_DELIVERED"] = formatEnergy(units_heat, 500.4*6.0*dt)

		return formatted
	}
}

func formatVal(units_temp, units_heat, key string, val dx2w.Value) string {
	switch key {

	case "OUTSIDE_AIR_TEMP":
		return formatTemperature(units_temp, val.Float32)

	case "DEW_POINT":
		return formatTemperature(units_temp, val.Float32)

	case "HP_INPUT_KW":
		return fmt.Sprintf("%.2f", val.Float32)

	case "NET_COP":
		return fmt.Sprintf("%.1f", val.Float32)

	case "LL_PRESSURE":
		return fmt.Sprintf("%.1f psi", val.Float32)

	case "LL_TEMP":
		return formatTemperature(units_temp, val.Float32)

	case "LIQUID_SUB-COOLING":
		return formatTemperature(units_temp, val.Float32)

	case "HP_WATER_DELTA-T":
		if val.Float32 == 0 {
			return "--"
		}
		return formatTemperature(units_temp, val.Float32)

	case "HP_OUTPUT_KW":
		return fmt.Sprintf("%.2fkW", val.Float32)

	case "HP_EXITING_WATER_TEMP":
		return formatTemperature(units_temp, val.Float32)

	case "HP_ENTERING_WATER_TEMP":
		return formatTemperature(units_temp, val.Float32)

	case "BUFFER_FLOW":
		return fmt.Sprintf("%.1f gpm", val.Float32)

	case "BUFFER_TANK_SETPOINT":
		return formatTemperature(units_temp, val.Float32)

	case "BUFFER_TANK_TEMP":
		return formatTemperature(units_temp, val.Float32)

	case "DIVERSION_VALVE_%_CLOSED":
		return fmt.Sprintf("%.0f %%", val.Float32)

	case "AUX_BOILER_KW":
		return fmt.Sprintf("%.2fkW", val.Float32)

	case "MIX_WATER_TEMP":
		return formatTemperature(units_temp, val.Float32)

	case "RETURN_WATER_TEMP":
		return formatTemperature(units_temp, val.Float32)

	default:
		if val.Type == "BOOL" {
			if val.Bool == true {
				return "true"
			}
			return "false"
		}
		return "NOIMP"
	}
}

func formatTemperature(units string, val float32) string {
	if strings.ToLower(units) == "celcius" {
		val = (val - 32.0) * 5.0 / 9.0
		return fmt.Sprintf("%.1f℃", val)
	}
	return fmt.Sprintf("%.1f℉", val)
}

func formatEnergy(units string, val float32) string {
	if strings.ToLower(units) == "btu" {
		return fmt.Sprintf("%.0f btu", val)
	}
	val = val / 3.412
	return fmt.Sprintf("%.2f kW", val)
}
