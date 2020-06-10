
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
		scriptEngine.GetHEPNodeID()
		scriptEngine.GetRawMessage()
		scriptEngine.SetRawMessage(value string)
		scriptEngine.SetCustomSIPHeader(map luatable)
		scriptEngine.SetHEPField(field string, value string)
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

	scriptEngine.SetCustomSIPHeader(headers)

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

	-- get the parsed SIP struct
	local sip = scriptEngine.GetSIPStruct()

	-- a struct can be nil so better check it
	if (sip == nil or sip == '') then
		return
	end

	if sip.FromHost == "127.0.0.1" then
		-- scriptEngine.Logp("ERROR", "found User-Agent:", sip.UserAgent)
	end

	scriptEngine.SetSIPHeader("FromHost", "1.1.1.1")

	-- Full SIP messsage can be changed
	-- scriptEngine.SetRawMessage("SIP 2/0")

	return 

end

-- this function will be executed third
function changeNodeIDtoNameFast()

	-- get only nodeID
	local nodeID = scriptEngine.GetHEPNodeID()
	if nodeID == 0 then
		scriptEngine.SetHEPField("NodeName","TestNode")
	end

	return 

end

-- this function will be executed fourth
function changeNodeIDtoNameSlow()
	-- get the parsed HEP struct
	local hep = scriptEngine.GetHEPStruct()

	-- a struct can be nil so better check it
	if (hep == nil or hep == '') then
		return
	end

	if hep.NodeID == 0 then
		hep.NodeName="TestNode"
	end

	return 

end


