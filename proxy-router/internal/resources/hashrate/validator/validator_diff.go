package validator

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
)

/* The meat... */

func ValidateDiff(en1 string, en2_size uint, job_diff uint64, version_mask string,
	job *stratumv1_message.MiningNotify, submit *stratumv1_message.MiningSubmit) (uint64, bool) {
	var prev_hash string
	var gen1 string
	var gen2 string
	var version string
	var en2 string
	var ntime string
	var nonce string
	var sver string
	var nbits string
	merkle_branches := []string{}

	/* extract job params */
	_ = json.Unmarshal(job.Params[1], &prev_hash)
	_ = json.Unmarshal(job.Params[2], &gen1)
	_ = json.Unmarshal(job.Params[3], &gen2)
	_ = json.Unmarshal(job.Params[4], &merkle_branches)
	_ = json.Unmarshal(job.Params[5], &version)
	_ = json.Unmarshal(job.Params[6], &nbits)

	/* extract submit params */
	en2 = submit.Params[2]
	ntime = submit.Params[3]
	nonce = submit.Params[4]

	/* pack header and hash */
	var header = new(bytes.Buffer)
	header.Grow(80)
	var gen = gen1 + en1 + en2 + gen2
	var merkle_root = sha256d(decode(gen))
	for _, branch := range merkle_branches {
		merkle_root = sha256d(append(merkle_root[:], decode(branch)...))
	}
	if len(submit.Params) > 5 {
		sver = submit.Params[5]
		jv := binary.LittleEndian.Uint32(decode_swap(version))
		sv := binary.LittleEndian.Uint32(decode_swap(sver))
		vm := binary.LittleEndian.Uint32(decode_swap(version_mask))
		nv := ((jv & ^vm) | (sv & vm))
		_ = binary.Write(header, binary.LittleEndian, nv)
	} else {
		_, _ = header.Write(decode_swap(version))
	}
	header.Write(decode_swap_words(prev_hash))
	header.Write(merkle_root[:])
	header.Write(decode_swap(ntime))
	header.Write(decode_swap(nbits))
	header.Write(decode_swap(nonce))
	hash := sha256d(header.Bytes())

	/* check result */
	h := new(big.Int)
	b := new(big.Int)
	t := new(big.Int)
	h.SetBytes(reverse(hash[:]))
	b.SetUint64(0xffff)
	b.Lsh(b, 208)
	t.Div(b, h)
	return t.Uint64(), t.Uint64() >= job_diff
}

func reverse(bytes []byte) []byte {
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return bytes
}

func decode(s string) []byte {
	res, _ := hex.DecodeString(s)
	return res
}

func decode_swap(s string) []byte {
	res, _ := hex.DecodeString(s)
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}
	return res
}

func decode_swap_words(s string) []byte {
	res, _ := hex.DecodeString(s)
	for w := 0; w < len(res); w += 4 {
		for i, j := w, w+4-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	}
	return res
}

func sha256d(data []byte) [32]byte {
	sum := sha256.Sum256(data)
	return sha256.Sum256(sum[:])
}
