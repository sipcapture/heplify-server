
-- this function will be executed first
function init()

	local protoType = HEP.api.getHEPProtoType()

	-- Check if we have SIP payload 
	if protoType ~= 1 then
		return
	end

	-- get All SIP parsed variables FromUser, ToUser, CallID
	local variables = HEP.api.getParsedVariables()
	-- original SIP message Payload
	local raw = HEP.api.getRawMessage()

	--[[ "applyHeader":        dec.applyHeader,
		"logData":            dec.logData,
		"setCustomHeaders":   dec.setCustomHeaders,
		"getParsedVariables": dec.getParsedVariables,
		"getHEPProtoType":    dec.getHEPProtoType,
		"getHEPSrcIP":        dec.getHEPSrcIP,
		"getHEPSrcPort":      dec.getHEPSrcPort,
		"getHEPDstIP":        dec.getHEPDstIP,
		"getHEPDstPort":      dec.getHEPDstPort,
		"getHEPTimeSeconds":  dec.getHEPTimeSeconds,
		"getHEPTimeUseconds": dec.getHEPTimeUseconds,
		"getHEPObject":       dec.getHEPObject,
		"getRawMessage":      dec.getRawMessage,
		"print":              fmt.Println,
	--]]

	-- HEP.api.logData("ERROR", "check:", raw)
	-- HEP.api.logData("DEBUG", "data", variables)

	-- Create lua table 
	local headers = {}
	headers["X-test"] = "Super TEST Header"

	-- local _, _, name, value = string.find(raw, "(Call-ID:)%s*:%s*(.+)")
	local name, value = raw:match("(CSeq):%s+(.-)\n")

	if name == "CSeq" then
		headers[name] = value
	end


	HEP.api.setCustomHeaders(headers)
	
	-- You can change any header and value . I.e. FromUser, "tester", X-CID
	-- Or replace full SIP messsage (RAW)
	-- HEP.api.applyHeader("RAW", "SIP 2/0")

	return 

end
