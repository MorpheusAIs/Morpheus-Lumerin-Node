package validator

import (
	sm "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
)

type testData struct {
	notify      *sm.MiningNotify
	submit1     *sm.MiningSubmit
	submit2     *sm.MiningSubmit
	xnonce      string
	xnonce2size int
	vmask       string
	diff        float64
}

func GetTestMsg() *testData {
	notify_raw := `{"id":null,"method":"mining.notify","params":["2dc3427c2e","221a7d5aeda279d8b8455fe56c8dc7d05582575d00038fbf0000000000000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4b03e1360cfabe6d6ddecabad1af6410018e1f62f26730ccb9c8a4a55c1c90fb96d7b124a68f126bcf0100000000000000","2e7c42c32d2f736c7573682f000000000383d02826000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac00000000000000002c6a4c2952534b424c4f434b3aa126fd3abcfed0d9d2fdf56d5650fda514e1a35408b1b8445c907d21005402510000000000000000266a24aa21a9ed217bdf1fc8e2ca2f98f2f3dc804fa19609ad045e8761e3fcd6b60baf80d1f5bf00000000",["fd90b0aa15698f631ae06aba1d688db974c899389c874f03b2c91784733ac50c","278cbb17943d36be5e7eaa08b70b733edd2fc6e4143ee7c184d63c8dcb22c48e","e02743f1b8d9050160c811cd8bec5af39c07a47ccade2a466e09409eeeb90b3b","3e76f16fc336d11a98b2c438e7c47b0a6c478a0f8df7c340f57db6887aa05a17","3d84d3378f3647157355aecf67c965f01729b3223f5a787aa7a3eb3de3a33e38","356c1febe5995abebfe7b4476efbf2b83158ce37757e3045f601bb1007ef9602","0064ab04971d60a5761ef07e7ff066a2ecfced7101e73a2d8cea07a730b50695","29ef3189c3aef0da8c1b7204a91f2384b73bdafcc7c1f9123a68b15316f7c5f8","bcc9dc862a6024d4f59ecb69f2adaa1911c6e4813d65a48ead8aa5e4151fb255","7accfd2b86edba50aa41fba73248ea365deaabf12a71c6e4b663be7438d9c091","4cc849c0f0d18f993634ba7563a6c68406e60e0ab1ac4f23bd18401b6e3ab7c4","9a6c6c8936fa3b807e3fbfc107c7427691246aecb89dd8e22bde58b7f748d21b"],"20000004","17056102","64c25820",false]}`
	submit_raw := `{"params": ["printcrypto.S19xp134tx6y164", "2dc3427c2e", "0a00000000000000", "64c25820", "591d28da", "00092000"], "id": 2955, "method": "mining.submit"}`
	submit_raw_2 := `{"params": ["printcrypto.S19xp134tx6y164", "2dc3427c2e", "0a00000000000000", "64c25820", "b6fda90f", "000b6000"], "id": 2956, "method": "mining.submit"}`
	xnonce := "11650804a6c84c"
	xnonce2size := 8
	vmask := "1fffe000"
	diff := 699.0

	notify, err := sm.ParseMiningNotify([]byte(notify_raw))
	if err != nil {
		panic(err)
	}

	submit1, err := sm.ParseMiningSubmit([]byte(submit_raw))
	if err != nil {
		panic(err)
	}

	submit2, err := sm.ParseMiningSubmit([]byte(submit_raw_2))
	if err != nil {
		panic(err)
	}

	return &testData{
		notify:      notify,
		submit1:     submit1,
		submit2:     submit2,
		xnonce:      xnonce,
		xnonce2size: xnonce2size,
		vmask:       vmask,
		diff:        diff,
	}
}
