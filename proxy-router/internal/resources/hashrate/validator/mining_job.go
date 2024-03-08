package validator

import (
	"encoding/hex"
	"sync"
	"time"

	sm "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
)

type shareBytes = [20]byte

type MiningJob struct {
	notify          *sm.MiningNotify
	diff            float64
	extraNonce1     string
	extraNonce2Size int
	// TODO: a quick fix of race condition in CheckDuplicateAndAddShare.
	// Sync map should not be needed here, because
	// all methods should be called from single goroutine, but as
	shares         sync.Map // map[shareBytes]bool
	expirationTime time.Time
}

func NewMiningJob(msg *sm.MiningNotify, diff float64, extraNonce1 string, extraNonce2Size int) *MiningJob {
	return &MiningJob{
		notify:          msg,
		diff:            diff,
		extraNonce1:     extraNonce1,
		extraNonce2Size: extraNonce2Size,
		shares:          sync.Map{},
	}
}

func (m *MiningJob) CheckDuplicateAndAddShare(s *sm.MiningSubmit) bool {
	bytes := SerializeShare(s.GetExtraNonce2(), s.GetNtime(), s.GetNonce(), s.GetVmask())
	_, loaded := m.shares.LoadOrStore(bytes, true)
	return loaded
}

func (m *MiningJob) GetNotify() *sm.MiningNotify {
	return m.notify.Copy()
}

func (m *MiningJob) GetDiff() float64 {
	return m.diff
}

func (m *MiningJob) GetExtraNonce1() string {
	return m.extraNonce1
}

func (m *MiningJob) GetExtraNonce2Size() int {
	return m.extraNonce2Size
}

// SerializeShare serializes the share into a 20-byte array.
// It includes only the fields that are unique for each share per job per destination
func SerializeShare(enonce2, ntime, nonce, vmask string) shareBytes {
	var hash shareBytes

	enonce2Bytes, _ := hex.DecodeString(enonce2)
	ntimeBytes, _ := hex.DecodeString(ntime)
	nonceBytes, _ := hex.DecodeString(nonce)
	vmaskBytes, _ := hex.DecodeString(vmask)

	copy(hash[:8], enonce2Bytes[:8])
	copy(hash[8:12], ntimeBytes[:4])
	copy(hash[12:16], nonceBytes[:4])
	copy(hash[16:20], vmaskBytes[:4])

	return hash
}

func (m *MiningJob) Copy() *MiningJob {
	return &MiningJob{
		notify:          m.notify.Copy(),
		diff:            m.diff,
		extraNonce1:     m.extraNonce1,
		extraNonce2Size: m.extraNonce2Size,
	}
}
