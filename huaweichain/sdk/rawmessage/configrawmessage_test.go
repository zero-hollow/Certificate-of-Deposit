package rawmessage

import (
	"sort"
	"testing"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/sdk/config"
)

const (
	allocKey               = "alloc"
	allocCoordinatorNumKey = "coordinator_num"
	allocPeerNumKey        = "peer_num"
	allocTypeKey           = "type"
	blockBatchKey          = "block_batch"
	attachableKey          = "attachable"
	domainKey              = "domains"
)

func newZoneCfg(id string, z *common.ZoneConf, ds []string) (*common.ZoneProperties, error) {
	m := make(map[string]interface{})
	alloc := make(map[string]interface{})
	if z.Alloc != nil {
		alloc[allocCoordinatorNumKey] = int(z.Alloc.MaxCoordinatorNum)
		alloc[allocPeerNumKey] = int(z.Alloc.MaxPeerNum)
		if z.Alloc.Type == 0 {
			alloc[allocTypeKey] = "balance"
		}
	}
	m[allocKey] = alloc
	m[blockBatchKey] = int(z.BlockBatch)
	m[attachableKey] = z.Attachable
	dm := make([]interface{}, 0)
	for _, s := range ds {
		dm = append(dm, interface{}(s))
	}
	m[domainKey] = dm

	p, err := config.BuildProperties(m)
	if err != nil {
		return nil, err
	}
	return &common.ZoneProperties{Id: id, Properties: p}, nil
}

func TestCheckAddZoneCfg(t *testing.T) {
	z := &common.ZoneConf{
		BlockBatch: 5,
		Attachable: true,
		Alloc: &common.Allocator{
			Type:              0,
			MaxPeerNum:        5,
			MaxCoordinatorNum: 5,
			Alloc:             nil,
		},
	}
	domains := []string{"/domainA", "/domainB"}
	zp, err := newZoneCfg("zoneB", z, domains)
	if err != nil {
		t.Errorf("%+v", err)
	}
	zones := make([]*common.ZoneProperties, 0)
	zones = append(zones, zp)

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Id < zones[j].Id
	})
	add := &common.ZoneOp_Add{Zones: zones}
	op := &common.ZoneOp{Op: &common.ZoneOp_Add_{Add: add}}
	err = checkAddCfg(op.GetAdd())

	if err != nil {
		t.Errorf("%+v", err)
	}
}
