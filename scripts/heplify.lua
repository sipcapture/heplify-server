
-- this function will be executed first
function checkRAW()
	--[[ Following functions can be used:
		GetHEPStruct()
		GetSIPStruct()
		GetHEPProtoType()
		GetHEPSrcIP()
		GetHEPSrcPort()
		GetHEPDstIP()
		GetHEPDstPort()
		GetHEPTimeSeconds()
		GetHEPTimeUseconds()
		GetHEPNodeID()
		GetRawMessage()
		SetRawMessage(value string)
		SetCustomSIPHeader(map luatable)
		SetHEPField(field string, value string)
		SetSIPHeader(header string, value string)
		Logp(level string, message string, data interface{})
		Print(text string)
	--]]
	
	local protoType = GetHEPProtoType()

	-- Check if we have SIP type 
	if protoType ~= 1 then
		return
	end

	-- original SIP message Payload
	local raw = GetRawMessage()
	-- Logp("DEBUG", "raw", raw)

	-- Create lua table 
	local headers = {}
	headers["X-test"] = "Super TEST Header"

	-- local _, _, name, value = string.find(raw, "(Call-ID:)%s*:%s*(.+)")
	local name, value = raw:match("(CSeq):%s+(.-)\n")

	if name == "CSeq" then
		headers[name] = value
	end

	SetCustomSIPHeader(headers)

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
	local sip = GetSIPStruct()

	-- a struct can be nil so better check it
	if (sip == nil or sip == '') then
		return
	end

	if sip.FromHost == "127.0.0.1" then
		-- Logp("ERROR", "found User-Agent:", sip.UserAgent)
	end

	SetSIPHeader("FromHost", "1.1.1.1")

	-- Full SIP messsage can be changed
	-- SetRawMessage("SIP 2/0")

	return 

end

-- this function will be executed third
function changeNodeIDtoName()

	-- get only nodeID
	local nodeID = GetHEPNodeID()
	if nodeID == 0 then
		SetHEPField("NodeName","TestNode")
	end

	return 

end

-- this function will be executed fourth
function sha1SumToCID()
	
	local sum = Hash("673187ceafc579fab78cc84cb1077a3f@0.0.0.0", "sha1")
	SetHEPField("CID", sum)

	return 

end

