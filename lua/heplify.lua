
-- this function will be executed first
function init()
	--[[ Following functions can be used:
		HEP.applyHeader(header string, value string)
		HEP.setCustomHeaders(m *map[string]string)
		HEP.getSIPObject()
		HEP.getHEPProtoType()
		HEP.getHEPSrcIP()
		HEP.getHEPSrcPort()
		HEP.getHEPDstIP()
		HEP.getHEPDstPort()
		HEP.getHEPTimeSeconds()
		HEP.getHEPTimeUseconds()
		HEP.getHEPObject()
		HEP.getRawMessage()
		HEP.logData(level string, message string, data interface{})
		HEP.print(text string)
	--]]
	
	local protoType = HEP.getHEPProtoType()

	-- Check if we have SIP payload 
	if protoType ~= 1 then
		return
	end

	-- get All SIP parsed variables FromUser, ToUser, CallID
	local variables = HEP.getSIPObject()
	-- original SIP message Payload
	local raw = HEP.getRawMessage()



	-- HEP.logData("ERROR", "check:", raw)
	-- HEP.logData("DEBUG", "data", variables)

	-- Create lua table 
	local headers = {}
	headers["X-test"] = "Super TEST Header"

	-- local _, _, name, value = string.find(raw, "(Call-ID:)%s*:%s*(.+)")
	local name, value = raw:match("(CSeq):%s+(.-)\n")

	if name == "CSeq" then
		headers[name] = value
	end


	HEP.setCustomHeaders(headers)
	
	-- You can change any header and value . I.e. FromUser, "tester", X-CID
	-- Or replace full SIP messsage (RAW)
	-- HEP.applyHeader("RAW", "SIP 2/0")

	return 

end
