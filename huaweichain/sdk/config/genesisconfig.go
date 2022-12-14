/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package config

import (
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/consensus"
	"github.com/pkg/errors"
)

// GenesisConfig is the definition of GenesisConfig.
type GenesisConfig struct {
	Consenters   []*consensus.Consenter
	Organs       []*common.Organization
	GenesisBlock GenesisBlock
	Decrypt      func([]byte) ([]byte, error)
}

// NewGenesisConfig is used to new an instance of genesis config
func NewGenesisConfig(genesisConfigPath string,
	decrypts ...func(bytes []byte) ([]byte, error)) (*GenesisConfig, error) {
	v, errMsg := parseCfg(genesisConfigPath)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "parse config file error")
	}

	decrypt := func(bytes []byte) ([]byte, error) {
		return bytes, nil
	}
	if len(decrypts) > 0 {
		decrypt = decrypts[0]
	}

	genesisConfig := &GenesisConfig{Decrypt: decrypt}
	errMsg = unmarshalGenesisConfig(v, &genesisConfig)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "error unmarshaling config into struct")
	}
	return genesisConfig, nil
}

// AddConsenter is used to add consenter.
func (gc *GenesisConfig) AddConsenter(consenter *consensus.Consenter) {
	gc.Consenters = append(gc.Consenters, consenter)
}

// AddOrganization is used to add organization.
func (gc *GenesisConfig) AddOrganization(name string, admin []byte, root []byte, tls []byte) {
	organ := &common.Organization{
		Name:        name,
		RootCert:    root,
		AdminCert:   admin,
		TLSRootCert: tls,
	}
	gc.Organs = append(gc.Organs, organ)
}

// GenesisBlock is the definition of GenesisBlock.
type GenesisBlock struct {
	DBType             string
	OrgLimit           bool
	Consensus          Consensus
	Raft               Raft
	Flic               Flic
	Hotstuff           Hotstuff
	Solo               Solo
	Policy             Policy
	Organizations      []Organization
	Consenters         []Consenter
	Net                *Net
	BlockLimit         int64
	MinPlatformVersion string
	ApprovalNoCert     bool
}

// Consensus is the definition of Consensus common cfg.
type Consensus struct {
	InitBlockSeqCap       uint64
	InitPendingEntriesCap uint64
	TickBufSize           uint64
	Type                  string
	BlockLimitRate        uint64
	TxLimitRate           uint64
	BlockPackage          *BlockPackage
	Pipeline              *Pipeline
	Transport             *Transport
}

// BlockPackage is the definition of Consensus module cfg.
type BlockPackage struct {
	MaxBlockSize uint64
	MaxTxCount   uint64
	TickInterval uint64
	TimeoutTick  uint64
	TxBufSize    uint64
	MaxRoutines  uint64
}

// Pipeline is the definition of pipeline.
type Pipeline struct {
	CommittedBlockBufSize uint64
	PersistedBlockBufSize uint64
	ProposedBlockBufSize  uint64
}

// Transport is the definition of consensus node send and receive msg buf cfg.
type Transport struct {
	SendBufSize uint64
	RecvBufSize uint64
}

// Hotstuff is the definition of Hotstuff consensus.
type Hotstuff struct {
	Nodes           []string
	Learners        []string
	ForwardProp     bool   `mapstructure:"forward_prop"`
	ProposerType    string `mapstructure:"proposer_type"`
	TickInterval    uint64 `mapstructure:"tick_interval"`
	ViewTimeoutTick uint32 `mapstructure:"view_timeout_tick"`
	MaxPrunedBlocks uint64 `mapstructure:"max_pruned_blocks"`
	MaxBlockBatch   uint64 `mapstructure:"max_block_batch"`
	MemoryStore     bool   `mapstructure:"memory_store"`
	VerifyTimestamp bool   `mapstructure:"verify_timestamp"`
	WaitTxDuration  uint64 `mapstructure:"wait_tx_duration"`
}

// Raft is the definition of Raft consensus.
type Raft struct {
	Group                       string
	Nodes                       []string
	Learners                    []string
	SafeTicker                  bool
	TickInterval                uint64
	HeartbeatTick               uint64
	ElectionTick                uint64
	UnreachableMissingHeartbeat uint64
	SnapshotThreshold           uint64
	CheckQuorum                 bool
	PreVote                     bool
	MaxBatchSize                uint64
	DisableForward              bool
	TermRoundBound              uint32
}

// Flic is the definition of Flic consensus.
type Flic struct {
	MaxFaultNodes uint64
	ReqTimeout    uint64
}

// Solo is the definition of Solo consensus.
type Solo struct {
}

// Policy is the definition of proposal policy.
type Policy struct {
	Config    string
	Lifecycle string
}

// Organization is the definition of Organization.
type Organization struct {
	ID          string // defined to store key when convert map to struct array
	AdminCert   string
	RootCert    string
	TLSRootCert string
}

// Consenter is the definition of consenter node.
type Consenter struct {
	ID       string // defined to store key when convert map to struct array
	Org      string
	Host     string
	Port     uint64
	ReqPort  uint32 `mapstructure:"req_port"`
	RestPort uint32 `mapstructure:"rest_port"`
	ReeCert  string
	TeeCert  string
}

// Net is the definition of node network.
type Net struct {
	Sync         Sync
	ZoneTemplate ZoneTemplate           `mapstructure:"zone_template"`
	AutoGenZone  bool                   `mapstructure:"auto_gen_zone"`
	ConsZone     map[string]interface{} `mapstructure:"cons_zone"`
	Zones        []map[string]interface{}
	Domains      []Domain
}

// Sync is the definition of sync.
type Sync struct {
	TickInterval uint64 `mapstructure:"tick_interval"`
	Heartbeat    uint64
	Timeout      uint64
}

// ZoneTemplate  is the definition of zone template.
type ZoneTemplate struct {
	Attachable bool
	BlockBatch uint64 `mapstructure:"block_batch"`
	Alloc      Allocator
}

// Allocator is the definition of allocator.
type Allocator struct {
	Type           string
	CoordinatorNum uint64 `mapstructure:"coordinator_num"`
	PeerNum        uint64 `mapstructure:"peer_num"`
	Conf           map[string]interface{}
}

// Zone is the definition of zone.
type Zone struct {
	ID         string
	BlockBatch uint64 `mapstructure:"block_batch"`
	Alloc      map[string]interface{}
	Domain     []string
	Attachable bool
}

// Domain is the definition of domain.
type Domain struct {
	Path string
}
