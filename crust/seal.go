package crust

import (
	"context"

	"github.com/crustio/go-ipfs-encryptor/utils"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

var sealBlackSet map[cid.Cid]bool
var sealBlackList = []string{
	"QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc",
	"QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn",
}

func init() {
	sealBlackSet = make(map[cid.Cid]bool)
	for _, v := range sealBlackList {
		c, _ := cid.Decode(v)
		sealBlackSet[c] = true
	}
}

func startSeal(root cid.Cid, value []byte, sessionKey string, sealedMap *map[cid.Cid]SealedBlock) (bool, error) {
	canSeal, path, err := sw.seal(root, sessionKey, false, value)
	if err != nil || !canSeal {
		return canSeal, err
	}

	sb := SealedBlock{
		Path: path,
		Size: len(value),
	}

	(*sealedMap)[root] = sb
	return true, nil
}

func sealBlock(root cid.Cid, leaf cid.Cid, value []byte, sessionKey string, sealedMap *map[cid.Cid]SealedBlock) (bool, error) {
	canSeal, path, err := sw.seal(root, sessionKey, false, value)
	if err != nil || !canSeal {
		return canSeal, err
	}

	sb := SealedBlock{
		Path: path,
		Size: len(value),
	}

	(*sealedMap)[leaf] = sb
	return true, nil
}

func endSeal(root cid.Cid, sessionKey string) (bool, error) {
	canSeal, _, err := sw.seal(root, sessionKey, false, []byte{})
	return canSeal, err
}

func deepSeal(ctx context.Context, originRootCid cid.Cid, rootNode ipld.Node, serv ipld.DAGService, sessionKey string, sealedMap *map[cid.Cid]SealedBlock) (bool, error) {
	for i := 0; i < len(rootNode.Links()); i++ {
		leafNode, err := serv.Get(ctx, rootNode.Links()[i].Cid)
		if err != nil {
			return false, err
		}

		canSeal, err := deepSeal(ctx, originRootCid, leafNode, serv, sessionKey, sealedMap)
		if err != nil || !canSeal {
			return canSeal, err
		}

		canSeal, err = sealBlock(originRootCid, leafNode.Cid(), leafNode.RawData(), sessionKey, sealedMap)
		if err != nil || !canSeal {
			return canSeal, err
		}
	}

	return true, nil
}

func Seal(ctx context.Context, root cid.Cid, serv ipld.DAGService) (bool, map[cid.Cid]SealedBlock, error) {
	// Black list
	if _, ok := sealBlackSet[root]; ok {
		return false, nil, nil
	}

	sealedMap := make(map[cid.Cid]SealedBlock)
	sessionKey := utils.RandStringRunes(8)
	rootNode, err := serv.Get(ctx, root)
	if err != nil {
		return false, nil, err
	}

	canSeal, err := startSeal(rootNode.Cid(), rootNode.RawData(), sessionKey, &sealedMap)
	if err != nil || !canSeal {
		return canSeal, nil, err
	}

	canSeal, err = deepSeal(ctx, rootNode.Cid(), rootNode, serv, sessionKey, &sealedMap)
	if err != nil || !canSeal {
		return canSeal, nil, err
	}

	canSeal, err = endSeal(root, sessionKey)
	return canSeal, sealedMap, err
}
