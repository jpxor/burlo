package main

// 1hr interval
var fields_static = []string{
	"HP_CT",
	"AUX_CT",
	"RADIANT_COOLING_MIN",
	"OUTDOOR_AIR_DESIGN_TEMP",
	"HOT_WATER_DESIGN_TEMP",
	"HOT_WATER_MAX_TEMP",
	"HOT_WATER_MIN_TEMP",
	"HOT_WATER_DIFFERENTIAL",
	"BOILER_CALL_DELAY_DURATION",
	"AUX_BOILER_BALANCE_POINT",
	"DIVERSION_VALVE_DIFFERENTIAL",
	"CHILLED_WATER_SETPOINT",
	"HP_OPERATING_MODE",
	"DEW_POINT_SAFETY_FACTOR",
	"EXWT_OFFSET",
	"RADIANT_COOLING_MODE",
	"FORCED_DEFROST",
	"BUFFER_TANK_DIFFERENTIAL",
	"BUFFER_FLOW",
	"HOT_WATER_TARGET",
}

// 10 min interval
var fields_slow = []string{
	"BUFFER_TANK_SETPOINT",
	"DIVERSION_VALVE_SETPOINT",
	"HP_KWH",
	"HP_OUTPUT_KWH",
	"AUX_KWH",
	"COMP_STALL_OR_DELAY_COUNTER",
}

// 1 min interval
var fields_1min_interval = []string{
	"BUFFER_TANK_TEMP",
	"COOLING_MODE",
	"HEATING_MODE",
	"COMPRESSOR_RUNTIME",
	"TIME_SINCE_LAST_DEFROST",
	"DEFROST",
	"OUTSIDE_AIR_TEMP",
	"DEW_POINT",
	"DIVERSION_VALVE_%_CLOSED",
}

// 15 sec interval
var fields_15sec_interval = []string{
	"LL_TEMP",
	"LL_PRESSURE",
	"LIQUID_SUB-COOLING",
	"HP_WATER_DELTA-T",
	"NET_COP",
	"HP_INPUT_KW",
	"HP_OUTPUT_KW",
	"AUX_BOILER_KW",
	"HP_ENTERING_WATER_TEMP",
	"HP_EXITING_WATER_TEMP",
	"MIX_WATER_TEMP",
	"RETURN_WATER_TEMP",
	"HP_CIRCULATOR",
	"COMPRESSOR_CALL",
}
