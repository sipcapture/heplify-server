package config

import "net/http"

func WebConfig(r *http.Request) HeplifyServer {
	var newSetting = Setting
	return newSetting
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
		
			<table>
			<tr>
				<td>HEPAddr</td>
				<td>HEPTLSAddr</td>
				<td>HEPTCPAddr</td>
			</tr>
			<tr>
				<td><input  type="text" name="HEPAddr" placeholder="{{.HEPAddr}}"></td>
				<td><input  type="text" name="HEPTLSAddr" placeholder="{{.HEPTLSAddr}}"></td>
				<td><input  type="text" name="HEPTCPAddr" placeholder="{{.HEPTCPAddr}}"></td>
			</tr>
			</table>

			<table>
			<tr>
				<td>ESAddr</td>
				<td>ESDiscovery</td>
			</tr>
			<tr>
				<td><input  type="text" name="ESAddr" placeholder="{{.ESAddr}}"></td>
				<td><input  type="text" name="ESDiscovery" placeholder="{{.ESDiscovery}}"></td>
			</tr>
			</table>

			<table>
			<tr>
				<td>LokiURL</td>
				<td>LokiBulk</td>
				<td>LokiTimer</td>
				<td>LokiBuffer</td>
			</tr>
			<tr>
				<td><input  type="text" name="LokiURL" placeholder="{{.LokiURL}}"></td>
				<td><input  type="text" name="LokiBulk" placeholder="{{.LokiBulk}}"></td>
				<td><input  type="text" name="LokiTimer" placeholder="{{.LokiTimer}}"></td>
				<td><input  type="text" name="LokiBuffer" placeholder="{{.LokiBuffer}}"></td>
			</tr>
			</table>

			<table>
			<tr>
				<td>PromAddr</td>
				<td>PromTargetIP</td>
				<td>PromTargetName</td>
			</tr>
			<tr>
				<td><input  type="text" name="PromAddr" placeholder="{{.PromAddr}}"></td>
				<td><input  type="text" name="PromTargetIP" placeholder="{{.PromTargetIP}}"></td>
				<td><input  type="text" name="PromTargetName" placeholder="{{.PromTargetName}}"></td>
			</tr>
			</table>

			
			<table>
			<tr>
				<td>DBShema</td>
				<td>DBDriver</td>
				<td>DBAddr</td>
				<td>DBUser</td>
				<td>DBPass</td>

			</tr>
			<tr>
				<td><input  type="text" name="DBShema" placeholder="{{.DBShema}}"></td>
				<td><input  type="text" name="DBDriver" placeholder="{{.DBDriver}}"></td>
				<td><input  type="text" name="DBAddr" placeholder="{{.DBAddr}}"></td>
				<td><input  type="text" name="DBUser" placeholder="{{.DBUser}}"></td>
				<td><input  type="text" name="DBPass" placeholder="{{"*******"}}"></td>
			</tr>
			</table>



			<input type="submit" value="Apply config" />
		</form>
    </body>
</html>
`
