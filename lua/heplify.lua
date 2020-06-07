
-- this function will be executed first
function checkRAW()
	--[[ Following functions can be used:
		scriptEngine.GetHEPStruct()
		scriptEngine.GetSIPStruct()
		scriptEngine.GetHEPProtoType()
		scriptEngine.GetHEPSrcIP()
		scriptEngine.GetHEPSrcPort()
		scriptEngine.GetHEPDstIP()
		scriptEngine.GetHEPDstPort()
		scriptEngine.GetHEPTimeSeconds()
		scriptEngine.GetHEPTimeUseconds()
		scriptEngine.GetRawMessage()
		scriptEngine.SetRawMessage(value string)
		scriptEngine.SetCustomHeader(map luatable)
		scriptEngine.SetSIPHeader(header string, value string)
		scriptEngine.Logp(level string, message string, data interface{})
		scriptEngine.Print(text string)
	--]]
	
	local protoType = scriptEngine.GetHEPProtoType()

	-- Check if we have SIP type 
	if protoType ~= 1 then
		return
	end

	-- original SIP message Payload
	local raw = scriptEngine.GetRawMessage()
	-- scriptEngine.Logp("DEBUG", "raw", raw)

	-- Create lua table 
	local headers = {}
	headers["X-test"] = "Super TEST Header"

	-- local _, _, name, value = string.find(raw, "(Call-ID:)%s*:%s*(.+)")
	local name, value = raw:match("(CSeq):%s+(.-)\n")

	if name == "CSeq" then
		headers[name] = value
	end

	scriptEngine.SetCustomHeader(headers)

	return 

end

-- this function will be executed second
function checkSIP()
	--[[ Following SIP header can be used:
		"ViaOne"
		"FromUser"
		"FromHost"
		"FromTag"
		"ToUser"
		"ToHost"
		"ToTag"
		"CallID"
		"XCallID"
		"ContactUser"
		"ContactHost"
		"Authorization.Username"
		"UserAgent"
		"Server"
		"PaiUser"
		"PaiHost"
	--]]

	-- get the parsed SIP object
	local sip = scriptEngine.GetSIPStruct()

	if sip.FromHost == "127.0.0.1" then
		-- scriptEngine.Logp("ERROR", "found User-Agent:", sip.UserAgent)
		print(sip.ToHost)
	end
	
	scriptEngine.SetSIPHeader("FromHost", "1.1.1.1")

	-- Full SIP messsage can be changed
	-- scriptEngine.SetRawMessage("SIP 2/0")

	return 

end