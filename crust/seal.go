package crust

import (
	"context"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

var sealedMap map[cid.Cid]map[cid.Cid]SealedBlock

var sealBlackSet map[cid.Cid]bool
var sealBlackList = []string{
	"QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc",
	"QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn",
}

func init() {
	sealedMap = make(map[cid.Cid]map[cid.Cid]SealedBlock)
	sealBlackSet = make(map[cid.Cid]bool)
	for _, v := range sealBlackList {
		c, _ := cid.Decode(v)
		sealBlackSet[c] = true
	}
}

func startSeal(root cid.Cid, value []byte) (bool, error) {
	sealedMap[root] = make(map[cid.Cid]SealedBlock)
	sb := SealedBlock{
		SHash: root.String(),
		Size:  len(value),
		Data:  value,
	}
	sealedMap[root][root] = sb
	return true, nil
}

func sealBlock(root cid.Cid, leaf cid.Cid, value []byte) error {
	sb := SealedBlock{
		SHash: leaf.String(),
		Size:  len(value),
		Data:  value,
	}
	sealedMap[root][leaf] = sb
	return nil
}

func endSeal(root cid.Cid) (map[cid.Cid]SealedBlock, error) {
	resMap := sealedMap[root]
	delete(sealedMap, root)
	return resMap, nil
}

func deepSeal(ctx context.Context, rootNode ipld.Node, serv ipld.DAGService) error {
	for i := 0; i < len(rootNode.Links()); i++ {
		leafNode, err := serv.Get(ctx, rootNode.Links()[i].Cid)
		if err != nil {
			return err
		}

		err = deepSeal(ctx, leafNode, serv)
		if err != nil {
			return err
		}

		err = sealBlock(rootNode.Cid(), leafNode.Cid(), leafNode.RawData())
		if err != nil {
			return err
		}
	}

	return nil
}

func Seal(ctx context.Context, root cid.Cid, serv ipld.DAGService) (bool, map[cid.Cid]SealedBlock, error) {
	// Black list
	if _, ok := sealBlackSet[root]; ok {
		return false, nil, nil
	}

	rootNode, err := serv.Get(ctx, root)
	if err != nil {
		return true, nil, err
	}

	needSeal, err := startSeal(rootNode.Cid(), rootNode.RawData())
	if !needSeal || err != nil {
		return needSeal, nil, err
	}

	err = deepSeal(ctx, rootNode, serv)
	if err != nil {
		return true, nil, err
	}

	sb, err := endSeal(root)
	return true, sb, err
}
