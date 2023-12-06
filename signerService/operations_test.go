package main

import (
	"encoding/json"
	"testing"
)

var tx = BasicTx{
	UserId:       "test",
	TokenId:      "ETH",
	BlockchainId: "ETH",
	AccountName:  "defaultAccount",
	Value:        "1",
	ToAddress:    "0xba536245A30404A983E120a3d07A7dF260a89669",
	FullTx:       "{\"type\":\"0x0\",\"nonce\":\"0xd\",\"gasPrice\":\"0x12c3045ef\",\"maxPriorityFeePerGas\":null,\"maxFeePerGas\":null,\"gas\":\"0x5208\",\"value\":\"0x2386f26fc10000\",\"input\":\"0x\",\"v\":\"0x0\",\"r\":\"0x0\",\"s\":\"0x0\",\"to\":\"0x019ad7b3a616275df4272adad98a95d07658789e\",\"hash\":\"0x115022a4912ff2b5c2e4f2fb9b22d2f38779fb795011b4fdc0bb5125e984ecef\"}",
	TxHash:       "0xcd2cbae4bd2be4a031042d65c1684a0b3795f0f4a5d6622d942a1a47e37e0f65",
	Status:       "test",
}

var hardcodedhash = "0xe4bffb33176924ea212a630405b4b44509d20cbaf152c463c88d273ffc37d683"

func TestPrepareETHTx(t *testing.T) {
	jsonbytes, err := json.Marshal(tx)
	if err != nil {
		t.Error("Error marshalling Basic Tx")
	}
	var ethTx BasicTx
	err = json.Unmarshal(jsonbytes, &ethTx)
	if err != nil {
		t.Error("Error unmarshalling Basic Tx")
	}

	if ethTx != tx {
		t.Error("JSON conversion error")
	}
}

func TestPrepareETHHash(t *testing.T) {
	_, err := prepareHash(hardcodedhash, "ETH")
	if err != nil {
		t.Error("Error calculating ETH hash")
	}
}
