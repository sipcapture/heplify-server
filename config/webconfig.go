package config

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

func WebConfig(r *http.Request) (*HeplifyServer, error) {
	var err error
	webSetting := Setting
	webSetting.HEPAddr = r.FormValue("HEPAddr")
	webSetting.HEPTCPAddr = r.FormValue("HEPTCPAddr")
	webSetting.HEPTLSAddr = r.FormValue("HEPTLSAddr")
	webSetting.ESAddr = r.FormValue("ESAddr")
	webSetting.LokiURL = r.FormValue("LokiURL")
	if webSetting.LokiBulk, err = strconv.Atoi(r.FormValue("LokiBulk")); err != nil {
		return nil, err
	}
	if webSetting.LokiTimer, err = strconv.Atoi(r.FormValue("LokiTimer")); err != nil {
		return nil, err
	}
	if webSetting.LokiBuffer, err = strconv.Atoi(r.FormValue("LokiBuffer")); err != nil {
		return nil, err
	}
	webSetting.DBShema = r.FormValue("DBShema")
	webSetting.DBDriver = r.FormValue("DBDriver")
	webSetting.DBAddr = r.FormValue("DBAddr")
	webSetting.DBUser = r.FormValue("DBUser")
	DBPass := r.FormValue("DBPass")
	if DBPass != "*******" {
		webSetting.DBPass = DBPass
	}
	webSetting.DBConfTable = r.FormValue("DBConfTable")
	if webSetting.DBBulk, err = strconv.Atoi(r.FormValue("DBBulk")); err != nil {
		return nil, err
	}
	if webSetting.DBTimer, err = strconv.Atoi(r.FormValue("DBTimer")); err != nil {
		return nil, err
	}
	if webSetting.DBBuffer, err = strconv.Atoi(r.FormValue("DBBuffer")); err != nil {
		return nil, err
	}
	if webSetting.DBWorker, err = strconv.Atoi(r.FormValue("DBWorker")); err != nil {
		return nil, err
	}
	DBRotate := r.FormValue("DBRotate")
	if DBRotate == "true" {
		webSetting.DBRotate = true
	} else if DBRotate == "false" {
		webSetting.DBRotate = false
	}
	if webSetting.DBDropDays, err = strconv.Atoi(r.FormValue("DBDropDays")); err != nil {
		return nil, err
	}
	if webSetting.DBDropDaysCall, err = strconv.Atoi(r.FormValue("DBDropDaysCall")); err != nil {
		return nil, err
	}
	if webSetting.DBDropDaysRegister, err = strconv.Atoi(r.FormValue("DBDropDaysRegister")); err != nil {
		return nil, err
	}
	if webSetting.DBDropDaysDefault, err = strconv.Atoi(r.FormValue("DBDropDaysDefault")); err != nil {
		return nil, err
	}
	Dedup := r.FormValue("Dedup")
	if Dedup == "true" {
		webSetting.Dedup = true
	} else if Dedup == "false" {
		webSetting.Dedup = false
	}

	if reflect.DeepEqual(webSetting, Setting) {
		return nil, fmt.Errorf("Equal config")
	}

	return &webSetting, nil
}

var WebForm = `
<!DOCTYPE html>
<html>
    <head>
		<title>heplify-server web config</title>
    </head>
    <body>
        <h2>heplify-server.toml</h2>
		<form method="POST">
		<style type="text/css">
		label {
			display: inline-block;
			width: 160px;
			text-align: left;
		}

		input[type=text], select {
			width: 15%;
			padding: 4px 6px;
			margin: 4px 0;
			display: inline-block;
			border: 1px solid #ccc;
			border-radius: 4px;
			box-sizing: border-box;
			text-align: right;
		  }

		  input[type=number], select {
			width: 15%;
			padding: 4px 6px;
			margin: 4px 0;
			display: inline-block;
			border: 1px solid #ccc;
			border-radius: 4px;
			box-sizing: border-box;
			text-align: right;
		  }
		  
		  input[type=submit] {
			width: 15%;
			background-color: #4CAF50;
			color: white;
			padding: 14px 20px;
			margin: 8px 0;
			border: none;
			border-radius: 4px;
			cursor: pointer;
		  }

		</style>

		<div>
			<input  type="text" name="HEPAddr" placeholder="{{.HEPAddr}}" value="{{.HEPAddr}}">
			<label>HEPAddr</label>
		</div>
		<div>
			<input  type="text" name="HEPTCPAddr" placeholder="{{.HEPTCPAddr}}" value="{{.HEPTCPAddr}}">
			<label>HEPTCPAddr</label>
		</div>
		<div>
			<input  type="text" name="HEPTLSAddr" placeholder="{{.HEPTLSAddr}}" value="{{.HEPTLSAddr}}">
			<label>HEPTLSAddr</label>
		</div>
		<div>
			<input  type="text" name="ESAddr" placeholder="{{.ESAddr}}" value="{{.ESAddr}}">
			<label>ESAddr</label>
		</div>
		<div>
			<input  type="text" name="LokiURL" placeholder="{{.LokiURL}}" value="{{.LokiURL}}">
			<label>LokiURL</label>
		</div>
		<div>
			<input  type="number" name="LokiBulk" placeholder="{{.LokiBulk}}" value="{{.LokiBulk}}" min="50" max="20000">
			<label>LokiBulk</label>
		</div>
		<div>
			<input  type="number" name="LokiTimer" placeholder="{{.LokiTimer}}" value="{{.LokiTimer}}" min="2" max="300">
			<label>LokiTimer</label>
		</div>
		<div>
			<input  type="number" name="LokiBuffer" placeholder="{{.LokiBuffer}}" value="{{.LokiBuffer}}" min="100" max="10000000">
			<label>LokiBuffer</label>
		</div>
		<div>
			<input  type="text" name="DBShema" placeholder="{{.DBShema}}" value="{{.DBShema}}">
			<label>DBShema</label>
		</div>
		<div>
			<input  type="text" name="DBDriver" placeholder="{{.DBDriver}}" value="{{.DBDriver}}">
			<label>DBDriver</label>
		</div>
		<div>
			<input  type="text" name="DBAddr" placeholder="{{.DBAddr}}" value="{{.DBAddr}}">
			<label>DBAddr</label>
		</div>
		<div>
			<input  type="text" name="DBUser" placeholder="{{.DBUser}}" value="{{.DBUser}}">
			<label>DBUser</label>
		</div>
		<div>
			<input  type="text" name="DBPass" placeholder="*******" value="*******">
			<label>DBPass</label>
		</div>
		<div>
			<input  type="number" name="DBBulk" placeholder="{{.DBBulk}}" value="{{.DBBulk}}" min="1" max="20000">
			<label>DBBulk</label>
		</div>
		<div>
			<input  type="number" name="DBTimer" placeholder="{{.DBTimer}}" value="{{.DBTimer}}" min="2" max="300">
			<label>DBTimer</label>
		</div>
		<div>
			<input  type="number" name="DBBuffer" placeholder="{{.DBBuffer}}" value="{{.DBBuffer}}" min="1000" max="10000000">
			<label>DBBuffer</label>
		</div>
		<div>
			<input  type="number" name="DBWorker" placeholder="{{.DBWorker}}" value="{{.DBWorker}}" min="1" max="40">
			<label>DBWorker</label>
		</div>
		<div>
			<input  type="text" name="DBRotate" placeholder="{{.DBRotate}}" value="{{.DBRotate}}">
			<label>DBRotate</label>
		</div>	
		<div>
			<input  type="number" name="DBDropDays" placeholder="{{.DBDropDays}}" value="{{.DBDropDays}}" min="0" max="3650">
			<label>DBDropDays</label>
		</div>	
		<div>
			<input  type="number" name="DBDropDaysCall" placeholder="{{.DBDropDaysCall}}" value="{{.DBDropDaysCall}}" min="0" max="3650">
			<label>DBDropDaysCall</label>
		</div>		
		<div>
			<input  type="number" name="DBDropDaysRegister" placeholder="{{.DBDropDaysRegister}}" value="{{.DBDropDaysRegister}}" min="0" max="3650">
			<label>DBDropDaysRegister</label>
		</div>
		<div>
			<input  type="number" name="DBDropDaysDefault" placeholder="{{.DBDropDaysDefault}}" value="{{.DBDropDaysDefault}}" min="0" max="3650">
			<label>DBDropDaysDefault</label>
		</div>
		<div>
			<input  type="text" name="DBConfTable" placeholder="{{.DBConfTable}}" value="{{.DBConfTable}}">
			<label>DBConfTable</label>
		</div>
		<div>
			<input  type="text" name="Dedup" placeholder="{{.Dedup}}" value="{{.Dedup}}">
			<label>Dedup</label>
		</div>

		</br><input type="submit" value="Apply config" />
		</form>
    </body>
</html>
`
