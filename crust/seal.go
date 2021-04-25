package crust

import (
	"context"
	"sync"

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

func startSeal(root cid.Cid, value []byte) (returnInfo, *SealedBlock) {
	canSeal, err := Worker.startSeal(root)
	if err != nil || !canSeal {
		return returnInfo{canSeal: canSeal, err: err}, nil
	}

	canSeal, path, err := Worker.seal(root, false, value)
	if err != nil || !canSeal {
		return returnInfo{canSeal: canSeal, err: err}, nil
	}

	sb := &SealedBlock{
		Path: path,
		Size: len(value),
	}

	return returnInfo{canSeal: true, err: nil}, sb
}

func sealBlockAsync(root cid.Cid, leaf cid.Cid, value []byte, serialMap *serialMap, lpool *utils.Lpool) {
	canSeal, path, err := Worker.seal(root, false, value)
	if err != nil || !canSeal {
		serialMap.rinfo <- returnInfo{canSeal: canSeal, err: err}
		lpool.Done()
		return
	}

	sb := SealedBlock{
		Path: path,
		Size: len(value),
	}

	serialMap.add(leaf, sb)
	lpool.Done()
}

func endSeal(root cid.Cid) returnInfo {
	canSeal, err := Worker.endSeal(root)
	return returnInfo{canSeal: canSeal, err: err}
}

func deepSeal(ctx context.Context, originRootCid cid.Cid, rootNode ipld.Node, serv ipld.DAGService, sealedMap *serialMap, lpool *utils.Lpool) returnInfo {
	for i := 0; i < len(rootNode.Links()); i++ {
		select {
		case rinfo := <-sealedMap.rinfo:
			return rinfo
		default:
		}

		leafNode, err := serv.Get(ctx, rootNode.Links()[i].Cid)
		if err != nil {
			return returnInfo{canSeal: false, err: err}
		}

		rinfo := deepSeal(ctx, originRootCid, leafNode, serv, sealedMap, lpool)
		if rinfo.err != nil || !rinfo.canSeal {
			return rinfo
		}

		lpool.Add(1)
		go sealBlockAsync(originRootCid, leafNode.Cid(), leafNode.RawData(), sealedMap, lpool)
	}

	return returnInfo{canSeal: true, err: nil}
}

func Seal(ctx context.Context, root cid.Cid, serv ipld.DAGService) (bool, map[cid.Cid]SealedBlock, error) {
	// Black list
	if _, ok := sealBlackSet[root]; ok {
		return false, nil, nil
	}

	// Base data
	sealedMap := newSerialMap()

	// Start seal root block
	rootNode, err := serv.Get(ctx, root)
	if err != nil {
		return false, nil, err
	}

	rinfo, sb := startSeal(rootNode.Cid(), rootNode.RawData())
	if rinfo.err != nil || !rinfo.canSeal {
		return rinfo.canSeal, nil, rinfo.err
	}
	sealedMap.data[rootNode.Cid()] = *sb

	// Multi-threaded deep seal
	lpool := utils.NewLpool(4)
	rinfo = deepSeal(ctx, rootNode.Cid(), rootNode, serv, sealedMap, lpool)
	if rinfo.err != nil || !rinfo.canSeal {
		return rinfo.canSeal, nil, rinfo.err
	}

	lpool.Wait()
	select {
	case rinfo := <-sealedMap.rinfo:
		return rinfo.canSeal, nil, rinfo.err
	default:
	}

	// End seal
	rinfo = endSeal(root)

	return rinfo.canSeal, sealedMap.data, rinfo.err
}

type returnInfo struct {
	canSeal bool
	err     error
}

// Convert map to serial map
type serialMap struct {
	data  map[cid.Cid]SealedBlock
	lock  sync.Mutex
	rinfo chan returnInfo
}

func newSerialMap() *serialMap {
	m := &serialMap{
		data:  make(map[cid.Cid]SealedBlock),
		rinfo: make(chan returnInfo, 16),
	}
	return m
}

func (m *serialMap) add(key cid.Cid, sb SealedBlock) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = sb
}
