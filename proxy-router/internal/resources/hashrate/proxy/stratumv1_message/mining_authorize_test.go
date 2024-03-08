package stratumv1_message

import (
	"fmt"
	"math/rand"
	"testing"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
)

var (
	id         = 1
	minerId    = "test-user"
	password   = "test-pwd"
	messageRaw = []byte(fmt.Sprintf(`{"id": %d, "method": "mining.authorize", "params": ["%s", "%s"]}`, id, minerId, password))
)

func TestNewMiningAuthorize(t *testing.T) {
	// creation
	authMsg := newMiningAuthorize(t)

	// getters
	if authMsg.GetID() != id {
		t.Fatalf("GetID")
	}
	if authMsg.GetUserName() != minerId {
		t.Fatalf("GetMinerID")
	}
	if authMsg.GetPassword() != password {
		t.Fatalf("GetPassword")
	}
}

func TestMinigAuthorizeSerialize(t *testing.T) {
	authMsg := newMiningAuthorize(t)

	serizalized := authMsg.Serialize()
	normalized, _ := lib.NormalizeJson(messageRaw)

	if string(normalized) != string(serizalized) {
		t.FailNow()
	}
}

func TestMiningAuthorizeSetters(t *testing.T) {
	authMsg := newMiningAuthorize(t)

	id = rand.Int()
	authMsg.SetID(id)
	if authMsg.GetID() != id {
		t.Fatalf("SetID")
	}

	minerId = "new-miner-id"
	authMsg.SetUserName(minerId)
	if authMsg.GetUserName() != minerId {
		t.Fatalf("SetMinerID")
	}

	password = "new-miner-pwd"
	authMsg.SetPassword(password)
	if authMsg.GetPassword() != password {
		t.Fatalf("SetPassword")
	}
}

func newMiningAuthorize(t *testing.T) *MiningAuthorize {
	msg, err := ParseMiningAuthorize(messageRaw)
	if err != nil {
		t.Fatal(err)
	}
	return msg
}

func TestMiningNotify(t *testing.T) {
	msg := []byte(`{"id":null,"method":"mining.notify","params":["6dd09b7f1","79fbbe417bf175f6ea1c3f0e5a58e63b55997ba20005fe540000000000000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4b035e5f0bfabe6d6d247a01c83f1e8158907b4b747694f04e1a816d8f5e4f0fee5dc1d749b13c274f0100000000000000","f1b709dd062f736c7573682f00000000034cd0a125000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac00000000000000002c6a4c2952534b424c4f434b3ab179eb054d678e491a395e5add9318347d32d199ed8986ba4452592400444e420000000000000000266a24aa21a9edd179590e017428082728d4793446d19e55ba2640fefeea047f7e9e5c4d2bc17f00000000",["82c86b00941dce4d7eda95656436bfac506a67eecca46cfebef69858dcd690ff","dbade76d91820f19ba5edc70f64f09c18afd97bbf6b5a0444042d5019d8da22c","5f9c47584149ecfeb3c0dad3e7078db17458fb93078aacbdc95da331d6e4f0fd","4b9b15494165ac9033a054921b625a84abe448d1be1870e60fb01ac6259cc57e","724b6632f170425d19b5d7389c1e83775d6e6a835e053598021cf52e39bb6440","ef9cc334314ce964df5f861e3a232c1173c19c862aa2134dbb0f5161a00e51c4","ec7b20a410ee8dc2cb2ded46da140827930318b40ac69951d7d59f5fe4d31b5a","1dbde9512f7fec9a33c4e9c0c6c4fef8d66cf0085c7c5319331293fba4fa76e6","149f72f88cb3bf81e5e1774d33841b716c642beb7caa34f3d8dc28d5e834fcea","bbd995247f4e4bbcaffc3c46e27db40854d5abac78bcce958bee7aa990b0094e"],"20000004","1709a7af","62d378ec",false]}`)
	obj, err := ParseMiningNotify(msg)
	if err != nil {
		panic(err)
	}

	msg2 := obj.Serialize()
	if string(msg2) != string(msg) {
		t.Fail()
	}
}

func TestMiningSubmit(t *testing.T) {
	b := `{"params": ["stage.s9x16", "00000000fc7d9b53", "c7520000000000", "62d8978d", "78563064"], "id": 2238, "method": "mining.submit"}`
	a, _ := ParseMiningSubmit([]byte(b))
	a.SetUserName("kiki")
	c := a.Serialize()
	fmt.Print(string(c))
}
