package crust

import (
	"encoding/json"
)

type SealedBlock struct {
	SHash string `json:"s_hash"`
	Size  int    `json:"size"`
	Data  []byte `json:"data"`
}

func TryGetSealedBlock(value []byte) (bool, *SealedBlock) {
	sb := &SealedBlock{}
	err := json.Unmarshal(value, sb)
	if err != nil {
		return false, nil
	}

	return true, sb
}

func (sb *SealedBlock) ToSealedInfo() *SealedInfo {
	return &SealedInfo{Sbs: []SealedBlock{*sb}}
}

type SealedInfo struct {
	Sbs []SealedBlock `json:"sbs"`
}

func (si *SealedInfo) Bytes() []byte {
	bs, _ := json.Marshal(si)
	return bs
}

func (si *SealedInfo) AddSealedBlock(sb SealedBlock) *SealedInfo {
	si.Sbs = append(si.Sbs, sb)
	return si
}

func TryGetSealedInfo(value []byte) (bool, *SealedInfo) {
	si := &SealedInfo{}
	err := json.Unmarshal(value, si)
	if err != nil {
		return false, nil
	}

	return true, si
}

func MergeSealedInfo(a *SealedInfo, b *SealedInfo) *SealedInfo {
	si := &SealedInfo{}
	si.Sbs = append(a.Sbs, b.Sbs...)
	return si
}
