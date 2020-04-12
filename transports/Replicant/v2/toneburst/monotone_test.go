package toneburst

//func TestMonotone_Perform(t *testing.T) {
//	thing1 := monolith.FixedByteType{Byte: 0xB0}
//	thing2 := monolith.FixedByteType{Byte: 0xB1}
//	items := []monolith.Monolith{thing1, thing2}
//	part := monolith.BytesPart{Items: items}
//	parts := []monolith.Monolith{part}
//	description := monolith.Description{Parts:parts}
//
//	removeSequences := []monolith.Description{description}
//
//	addInstance := monolith.Instance{
//		Desc: description,
//		Args: make([]interface{}, 0),
//	}
//	addSequences := []monolith.Instance{addInstance}
//	println("Add Sequences count: ", len(addSequences))
//
//	monotoneConfig := MonotoneConfig{
//		AddSequences:    addSequences,
//		RemoveSequences: removeSequences,
//		SpeakFirst:      true,
//	}
//
//	println("Add Sequences count: ", len(monotoneConfig.AddSequences))
//	monotoneInstance := NewMonotone(monotoneConfig)
//	go monotoneServer(monotoneInstance)
//
//	clientConn, dialError := net.Dial("tcp", "127.0.0.1:2121")
//	if dialError != nil {
//		println("Dial error: ")
//		println(dialError.Error())
//		t.Fail()
//	}
//
//	performError := monotoneInstance.Perform(clientConn)
//	if performError!= nil {
//		println("Error on perform: ")
//		println(performError.Error())
//		t.Fail()
//	}
//}
//
//func monotoneServer(monotoneInstance *Monotone) {
//	l, serverErr := net.Listen("tcp", "0.0.0.0:2121")
//	if serverErr != nil {
//		println(serverErr.Error())
//		return
//	}
//
//	serverConn, acceptError := l.Accept()
//	if acceptError != nil {
//		println(acceptError.Error())
//		return
//	}
//
//	serverPerformError := monotoneInstance.Perform(serverConn)
//	if serverPerformError!= nil {
//		println("Error on perform: ", serverPerformError)
//		return
//	}
//}
