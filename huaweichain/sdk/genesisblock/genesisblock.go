/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

// Package genesisblock is the definition of parse genesis block from genesis config file.
package genesisblock

import (
	"io/ioutil"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"git.huawei.com/huaweichain/common/logger"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/configuration"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/consensus"

	"git.huawei.com/huaweichain/sdk/config"
)

const (
	domainKey       = "domains"
	domainSeparator = "/"

	sep      = "."
	segCount = 2

	base    = 10
	bitSize = 64
)

var log = logger.GetModuleLogger("go sdk", "genesis block")

// GetChainConfig is used to parse chain config from config file.
func GetChainConfig(c *config.GenesisConfig, chainID string) (*common.ChainConfig, error) {
	if err := checkGenesisBlock(c); err != nil {
		return nil, err
	}
	dbType, err := consensusDBType(c)
	if err != nil {
		return nil, errors.WithMessage(err, "get DB type error")
	}
	sysConsensus, err := getConsensus(c, chainID, c.Decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get sys consensus error")
	}
	organs, err := getOrganizations(c, c.Decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get organizations error")
	}

	consensusByte, err := sysConsensus.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "chain config marshal error")
	}

	var networkBytes []byte
	if c.GenesisBlock.Net != nil {
		network, netErr := getNetWorkConfig(c)
		if netErr != nil {
			return nil, errors.WithMessage(netErr, "get network config error")
		}
		networkBytes, err = network.Marshal()
		if err != nil {
			return nil, errors.WithMessage(err, "network config marshal error")
		}
	}

	minPltfVersion, err := minPlatformVersion(c.GenesisBlock.MinPlatformVersion)
	if err != nil {
		return nil, errors.WithMessage(err, "min platform version is invalid")
	}

	chainConfig := &common.ChainConfig{
		ChainId:            chainID,
		DbType:             dbType,
		OrgLimit:           c.GenesisBlock.OrgLimit,
		Organizations:      organs,
		Consensus:          consensusByte,
		Network:            networkBytes,
		ConfigPolicy:       c.GenesisBlock.Policy.Config,
		LifecyclePolicy:    c.GenesisBlock.Policy.Lifecycle,
		BlockLimit:         c.GenesisBlock.BlockLimit,
		MinPlatformVersion: minPltfVersion,
		ApprovalNoCert:     c.GenesisBlock.ApprovalNoCert,
	}
	return chainConfig, nil
}

// ValidatePlatformVersion check if the given version string is a valid
// version string.
func ValidatePlatformVersion(ver string) error {
	if ver == "" {
		return errors.New("new min platform version of chain should be specified")
	}
	segments := strings.Split(ver, sep)
	if len(segments) != segCount {
		return errors.Errorf(
			"min platform version should contain 2 segments, but got %d", len(segments))
	}
	for _, seg := range segments {
		_, err := strconv.ParseUint(seg, base, bitSize)
		if err != nil {
			return errors.Errorf("segment %s is not numeric: %s", seg, err)
		}
	}
	return nil
}

func minPlatformVersion(ver string) (string, error) {
	if err := ValidatePlatformVersion(ver); err != nil {
		return "", errors.WithMessage(err, "validate min platform version")
	}
	return ver, nil
}

func consensusDBType(c *config.GenesisConfig) (common.DBType, error) {
	switch dbName := c.GenesisBlock.DBType; dbName {
	case "leveldb":
		return common.LEVELDB, nil
	case "mysql":
		return common.MYSQL, nil
	case "postgresql":
		return common.POSTGRESQL, nil
	case "rocksdb":
		return common.ROCKSDB, nil
	default:
		return 0, errors.Errorf("unknow DBType: %s", dbName)
	}
}

func getNetWorkConfig(genesisConfig *config.GenesisConfig) (*common.NetworkConfig, error) {
	nwc := &common.NetworkConfig{}
	setSync(genesisConfig.GenesisBlock.Net, nwc)
	if err := setZoneTemplate(genesisConfig.GenesisBlock.Net, nwc); err != nil {
		return nil, errors.WithMessage(err, "set zone template error")
	}
	nwc.AutoGenZone = genesisConfig.GenesisBlock.Net.AutoGenZone
	setDomains(genesisConfig.GenesisBlock.Net, nwc)
	if err := setConsZone(genesisConfig.GenesisBlock.Net, nwc); err != nil {
		return nil, errors.WithMessage(err, "set cons zone error")
	}
	if err := setZones(genesisConfig.GenesisBlock.Net, nwc); err != nil {
		return nil, errors.WithMessage(err, "set zones error")
	}
	if err := config.CheckSyncCfg(nwc); err != nil {
		return nil, err
	}
	return nwc, nil
}

func setDomains(net *config.Net, nwc *common.NetworkConfig) {
	allDomains := completeDomains(net.Domains)
	nwc.Domains = make([]*common.Domain, len(allDomains))
	nwc.Zones = make([]*common.Zone, 0)
	for i, domain := range allDomains {
		if net.AutoGenZone {
			path := domain.Path[1:]
			id := strings.ReplaceAll(path, domainSeparator, "::")
			zone := &common.Zone{
				Id:      id,
				Domains: []string{domain.Path},
				Conf:    nwc.ZoneTemplate,
				AutoGen: true,
			}
			nwc.Zones = append(nwc.Zones, zone)
		}
		nwc.Domains[i] = &common.Domain{Path: domain.Path}
	}
	sort.Slice(nwc.Domains, func(i, j int) bool {
		return nwc.Domains[i].Path < nwc.Domains[j].Path
	})
}

func completeDomains(domains []config.Domain) []config.Domain {
	dps := make([]string, len(domains))
	for idx, d := range domains {
		dps[idx] = d.Path
	}
	filteredDomainPaths := removePrefix(dps)
	m := make(map[string]interface{})
	for _, fdp := range filteredDomainPaths {
		fdp = fdp[1:]
		partPaths := strings.Split(fdp, domainSeparator)
		path := ""
		for _, part := range partPaths {
			path = path + domainSeparator + part
			m[path] = struct{}{}
		}
	}

	filteredDomains := make([]config.Domain, 0, len(m))
	for key := range m {
		filteredDomains = append(filteredDomains, config.Domain{Path: key})
	}
	return filteredDomains
}

func removePrefix(domains []string) []string {
	var filteredDomains []string
	for _, domain := range domains {
		if !isPrefix(domain, domains) {
			filteredDomains = append(filteredDomains, domain)
		}
	}
	return filteredDomains
}

func isPrefix(domain string, domains []string) bool {
	for _, d := range domains {
		if d != domain && strings.HasPrefix(d, domain) {
			return true
		}
	}
	return false
}

func setZones(net *config.Net, nwc *common.NetworkConfig) error {
	for _, zone := range net.Zones {
		values, ok := zone["id"]
		if !ok {
			return errors.New("zone id is missing")
		}
		var id string
		id, ok = values.(string)
		if !ok {
			return errors.Errorf("expected type: string, but real type: %v", reflect.TypeOf(values))
		}

		domains, err := parseDomains(zone)
		if err != nil {
			return errors.WithMessage(err, "parse domains error")
		}
		zoneConf := &common.ZoneConf{Alloc: &common.Allocator{}}
		err = config.SetZoneConf(zoneConf, zone, nwc.ZoneTemplate)
		if err != nil {
			return errors.WithMessage(err, "set zone conf error")
		}
		z := &common.Zone{
			Id:      id,
			Domains: domains,
			Conf:    zoneConf,
		}
		nwc.Zones = append(nwc.Zones, z)
	}
	sort.Slice(nwc.Zones, func(i, j int) bool {
		return nwc.Zones[i].Id < nwc.Zones[j].Id
	})
	return nil
}

func parseDomains(zone map[string]interface{}) ([]string, error) {
	var domains []string
	if zone == nil {
		return domains, errors.New("zone is empty")
	}
	value, ok := zone[domainKey]
	if ok {
		ds, ok := value.([]interface{})
		if !ok {
			return domains, errors.Errorf("expected type: string, but real type: %v", reflect.TypeOf(value))
		}
		for _, d := range ds {
			domain, ok := d.(string)
			if !ok {
				return domains, errors.Errorf("expected type: string, but real type: %v", reflect.TypeOf(d))
			}
			domains = append(domains, domain)
		}
	}
	return domains, nil
}

func setSync(net *config.Net, nwc *common.NetworkConfig) {
	nwc.Sync = &common.SyncConf{
		TickInterval:  uint32(net.Sync.TickInterval),
		HeartbeatTick: uint32(net.Sync.Heartbeat),
		TimeoutTick:   uint32(net.Sync.Timeout),
	}
}

func setZoneTemplate(net *config.Net, nwc *common.NetworkConfig) error {
	zoneConf, err := getZoneConfFromZoneTemplate(net.ZoneTemplate)
	if err != nil {
		return errors.WithMessage(err, "get zone conf from zone template error")
	}
	nwc.ZoneTemplate = zoneConf
	return nil
}

func setConsZone(net *config.Net, nwc *common.NetworkConfig) error {
	var err error
	nwc.ConsZone = &common.ZoneConf{Alloc: &common.Allocator{}}
	err = config.SetZoneConf(nwc.ConsZone, net.ConsZone, nwc.ZoneTemplate)
	if err != nil {
		return errors.WithMessage(err, "parse zone conf error")
	}
	return nil
}

func getZoneConfFromZoneTemplate(zoneTemplate config.ZoneTemplate) (*common.ZoneConf, error) {
	alloc, err := getAllocator(zoneTemplate.Alloc)
	if err != nil {
		return nil, errors.WithMessage(err, "get allocator error")
	}
	return &common.ZoneConf{
		BlockBatch: uint32(zoneTemplate.BlockBatch),
		Attachable: zoneTemplate.Attachable,
		Alloc:      alloc,
	}, nil
}

func getAllocator(a config.Allocator) (*common.Allocator, error) {
	allocType, ok := common.AllocType_value[strings.ToUpper(a.Type)]
	if !ok {
		return nil, errors.Errorf("invalid alloc type[%v]", a.Type)
	}
	alloc := &common.Allocator{
		Type:              common.AllocType(allocType),
		MaxCoordinatorNum: uint32(a.CoordinatorNum),
		MaxPeerNum:        uint32(a.PeerNum),
	}

	value, ok := a.Conf["fan"]
	if !ok {
		return nil, errors.New("no fan out cfg")
	}
	fan, ok := value.(int)
	if !ok {
		return nil, errors.Errorf("type assert error, expected type: %v, but real type: %v",
			reflect.TypeOf(fan), reflect.TypeOf(value))
	}

	alloc.Alloc = &common.Allocator_Balance{Balance: &common.BalanceAlloc{Fan: uint32(fan)}}
	return alloc, nil
}

func getConsensus(genesisConfig *config.GenesisConfig, chainID string,
	decrypt func(bytes []byte) ([]byte, error)) (*consensus.SysCons, error) {
	sysConsensus := &consensus.SysCons{}
	if err := setConsCommonCfg(&genesisConfig.GenesisBlock.Consensus, sysConsensus); err != nil {
		return nil, err
	}
	if len(genesisConfig.Consenters) > 0 {
		sysConsensus.Consenter = genesisConfig.Consenters
	} else if err := setConsenters(genesisConfig, sysConsensus, decrypt); err != nil {
		return nil, err
	}

	if err := setConsensusType(genesisConfig, chainID, sysConsensus, decrypt); err != nil {
		return nil, err
	}
	return sysConsensus, nil
}

func setConsensusType(genesisConfig *config.GenesisConfig, chainID string, sysConsensus *consensus.SysCons,
	decrypt func(bytes []byte) ([]byte, error)) error {
	consensusType := genesisConfig.GenesisBlock.Consensus.Type
	switch consensusType {
	case "solo":
		sysConsensus.Type = consensus.Solo
		sysConsensus.Genesis = &consensus.Genesis{Conf: &consensus.Genesis_Solo{Solo: &consensus.SoloConfig{}}}
	case "flic":
		flicIns := &consensus.Genesis_Flic{Flic: &consensus.FlicConfig{
			MaxFaultNodes: genesisConfig.GenesisBlock.Flic.MaxFaultNodes,
			ReqTimeout:    genesisConfig.GenesisBlock.Flic.ReqTimeout}}
		if err := checkFlicTypeCfg(flicIns.Flic); err != nil {
			return err
		}
		sysConsensus.Type = consensus.PBFT
		sysConsensus.Genesis = &consensus.Genesis{Conf: flicIns}
	case "raft":
		initialState, err := getInitialState(&genesisConfig.GenesisBlock, chainID, decrypt)

		if err != nil {
			return errors.WithMessage(err, "get initial state...failed")
		}

		sysConsensus.Genesis = &consensus.Genesis{Conf: &consensus.Genesis_Raft{Raft: &consensus.RaftConfig{
			InitialState: initialState,
		}}}
		sysConsensus.Type = consensus.Raft
	case "hotstuff":
		hotstuffCfg, err := getHotstuffCfg(genesisConfig, chainID, decrypt)
		if err != nil {
			return err
		}
		sysConsensus.Type = consensus.Hotstuff
		sysConsensus.Genesis = &consensus.Genesis{Conf: &consensus.Genesis_Hotstuff{Hotstuff: hotstuffCfg}}
	default:
		return errors.Errorf("not support consensus type: %v", consensusType)
	}
	return nil
}

func getHotstuffCfg(g *config.GenesisConfig, chainID string,
	decrypt func(bytes []byte) ([]byte, error)) (*consensus.HotStuffConfig, error) {
	genesis := &g.GenesisBlock
	h := genesis.Hotstuff

	consenters, err := getConsenters(genesis, h.Nodes, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get certs...failed")
	}

	learners, err := getLearners(h.Nodes, h.Learners)
	if err != nil {
		return nil, errors.WithMessage(err, "get learners...failed")
	}

	proposer, err := getProposerType(h.ProposerType)
	if err != nil {
		return nil, err
	}

	cfg := &consensus.HotStuffConfig{
		Epoch:           0,
		Group:           chainID,
		Consenters:      consenters,
		Learners:        learners,
		MaxPrunedBlocks: h.MaxPrunedBlocks,
		Tick:            h.TickInterval,
		ViewTimeout:     h.ViewTimeoutTick,
		Type:            proposer,
		ForwardProp:     h.ForwardProp,
		MaxBlockBatch:   h.MaxBlockBatch,
		MemoryStore:     h.MemoryStore,
		VerifyTimestamp: h.VerifyTimestamp,
		WaitTxDuration:  h.WaitTxDuration,
	}

	if err := checkHotstuffCfg(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getProposerType(t string) (consensus.ProposerType, error) {
	var proposer consensus.ProposerType
	switch t {
	case "", "fixed":
		proposer = consensus.FixedProposer
	case "view":
		proposer = consensus.ViewProposer
	default:
		return proposer, errors.Errorf("unsupported proposer type: %s", t)
	}

	return proposer, nil
}

func setConsCommonCfg(cfg *config.Consensus, sysCons *consensus.SysCons) error {
	c := &consensus.CommonConfig{
		MaxTxCount:            cfg.BlockPackage.MaxTxCount,
		MaxBlockSize:          cfg.BlockPackage.MaxBlockSize,
		TimeoutTick:           cfg.BlockPackage.TimeoutTick,
		TickInterval:          cfg.BlockPackage.TickInterval,
		TxBufSize:             cfg.BlockPackage.TxBufSize,
		MaxRoutines:           cfg.BlockPackage.MaxRoutines,
		InitBlockSeqCap:       cfg.InitBlockSeqCap,
		InitPendingEntriesCap: cfg.InitPendingEntriesCap,
		TickBufSize:           cfg.TickBufSize,
		ProposedBlockBufSize:  cfg.Pipeline.ProposedBlockBufSize,
		CommittedBlockBufSize: cfg.Pipeline.CommittedBlockBufSize,
		PersistedBlockBufSize: cfg.Pipeline.PersistedBlockBufSize,
		BlockLimitRate:        cfg.BlockLimitRate,
		TxLimitRate:           cfg.TxLimitRate,
		SendBufSize:           cfg.Transport.SendBufSize,
		RecvBufSize:           cfg.Transport.RecvBufSize,
	}
	if err := checkCommonCfg(c); err != nil {
		return err
	}
	sysCons.Conf = c
	return nil
}

func setConsenters(c *config.GenesisConfig, sysConsensus *consensus.SysCons,
	decrypt func(bytes []byte) ([]byte, error)) error {
	var consenters []*consensus.Consenter
	for idx := range c.GenesisBlock.Consenters {
		cn := consenter(c.GenesisBlock.Consenters[idx])
		reeCert, err := ioutil.ReadFile(cn.ReeCert)
		if err != nil {
			return errors.WithMessage(err, "load ree certificate error")
		}
		reeCert, err = decrypt(reeCert)
		if err != nil {
			return errors.WithMessage(err, "decrypt message error")
		}
		teeCert, err := ioutil.ReadFile(cn.TeeCert)
		if err != nil {
			return errors.WithMessage(err, "load tee certificate error")
		}
		teeCert, err = decrypt(teeCert)
		if err != nil {
			return errors.WithMessage(err, "decrypt message error")
		}
		consenters = append(consenters, cn.conv(reeCert, teeCert))
	}
	sysConsensus.Consenter = consenters
	return nil
}

func isContain(nodes []string, learner string) bool {
	for _, n := range nodes {
		if n == learner {
			return true
		}
	}
	return false
}

// getLearners for raft and hotstuff consensus
func getLearners(ns []string, ls []string) ([]string, error) {
	var learners []string
	for _, l := range ls {
		if isContain(ns, l) {
			learners = append(learners, l)
		} else {
			return nil, errors.Errorf("get learner name from nodes...failed! name: %s", l)
		}
	}
	return learners, nil
}

// getConsenters for raft and hotstuff consensus
func getConsenters(b *config.GenesisBlock, nodes []string,
	decrypt func(bytes []byte) ([]byte, error)) ([]*consensus.Consenter,
	error) {
	var consenters []*consensus.Consenter
	for _, name := range nodes {
		for i := 0; i < len(b.Consenters); i++ {
			c := consenter(b.Consenters[i])
			if c.ID != name {
				continue
			}
			reeCert, err := ioutil.ReadFile(c.ReeCert)
			if err != nil {
				return nil, errors.WithMessage(err, "load ree certificate error")
			}
			teeCert, err := ioutil.ReadFile(c.TeeCert)
			if err != nil {
				log.Warnf("failed to read tee cert from file because %+v, ignore and replaced by ree cert.",
					errors.WithMessage(err, "load tee certificate error"))
				teeCert = reeCert
			}
			reeCert, err = decrypt(reeCert)
			if err != nil {
				return nil, errors.WithMessage(err, "decrypt message error")
			}
			teeCert, err = decrypt(teeCert)
			if err != nil {
				return nil, errors.WithMessage(err, "decrypt message error")
			}
			consenters = append(consenters, c.conv(reeCert, teeCert))
			break
		}
	}
	return consenters, nil
}

// getInitialState for raft consensus
func getInitialState(b *config.GenesisBlock, chainID string,
	decrypt func(bytes []byte) ([]byte, error)) ([]byte, error) {
	learners, err := getLearners(b.Raft.Nodes, b.Raft.Learners)
	if err != nil {
		return nil, errors.WithMessage(err, "get certs...failed")
	}
	consenters, err := getConsenters(b, b.Raft.Nodes, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get certs...failed")
	}
	state := &consensus.InitialState{
		Group:          chainID,
		Consenters:     consenters,
		Learners:       learners,
		Safe:           b.Raft.SafeTicker,
		Tick:           uint32(b.Raft.TickInterval),
		Heartbeat:      uint32(b.Raft.HeartbeatTick),
		Election:       uint32(b.Raft.ElectionTick),
		Unreachable:    uint32(b.Raft.UnreachableMissingHeartbeat),
		SnapThreshold:  b.Raft.SnapshotThreshold,
		CheckQuorum:    b.Raft.CheckQuorum,
		PreVote:        b.Raft.PreVote,
		MaxBatchSize:   b.Raft.MaxBatchSize,
		DisableForward: b.Raft.DisableForward,
		TermRoundBound: b.Raft.TermRoundBound,
	}
	if err = checkInitStateCfg(state); err != nil {
		return nil, err
	}
	stateByte, err := state.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "chain config marshal error")
	}
	return stateByte, nil
}

func getOrganizations(c *config.GenesisConfig,
	decrypt func(bytes []byte) ([]byte, error)) ([]*common.Organization, error) {
	if len(c.Organs) > 0 {
		return c.Organs, nil
	}
	var organs []*common.Organization
	for idx := range c.GenesisBlock.Organizations {
		organNode := c.GenesisBlock.Organizations[idx]
		rootCert, err := ioutil.ReadFile(organNode.RootCert)
		if err != nil {
			return nil, errors.WithMessage(err, "load root certifacate error")
		}
		rootCert, err = decrypt(rootCert)
		if err != nil {
			return nil, errors.WithMessage(err, "decrypt message error")
		}
		adminCert, err := ioutil.ReadFile(organNode.AdminCert)
		if err != nil {
			return nil, errors.WithMessage(err, "load admin certifacate error")
		}
		adminCert, err = decrypt(adminCert)
		if err != nil {
			return nil, errors.WithMessage(err, "decrypt message error")
		}
		tlsRootCert, err := ioutil.ReadFile(organNode.TLSRootCert)
		if err != nil {
			return nil, errors.WithMessage(err, "load tls root cert error")
		}
		tlsRootCert, err = decrypt(tlsRootCert)
		if err != nil {
			return nil, errors.WithMessage(err, "decrypt message error")
		}

		organ := &common.Organization{Name: organNode.ID, RootCert: rootCert,
			AdminCert: adminCert, TLSRootCert: tlsRootCert}
		organs = append(organs, organ)
	}
	return organs, nil
}

func checkCfgList(cl *configuration.List) error {
	if err := cl.CheckCfgList(&cl.DefaultCfgLogs, &cl.ErrorCfgLogs); err != nil {
		return err
	}
	if cnt, logStr := cl.DefaultCfgLogs.LogStr(); cnt != 0 {
		log.Infof("[NAME]%s:\n%s\n[TOTAL] %v configs...use default value\n", cl.Name, logStr, cnt)
	}
	if cnt, logStr := cl.ErrorCfgLogs.LogStr(); cnt != 0 {
		return errors.Errorf("\n[NAME]%s:\n%s\n[ERROR] check config param...failed", cl.Name, logStr)
	}
	return nil
}

// check param
func checkGenesisBlock(c *config.GenesisConfig) error {
	entries := []configuration.Entry{
		{"ConsensusType", &c.GenesisBlock.Consensus.Type, "solo", nil, nil},
		{"DBType", &c.GenesisBlock.DBType, "leveldb", nil, nil},
	}
	list := configuration.List{Name: "GenesisBlock", Entries: entries}

	if err := checkCfgList(&list); err != nil {
		return err
	}
	return nil
}

const (
	defaultMaxTxCount            = 500
	defaultMaxBlockSize          = 1024
	defaultTickInterval          = 100
	defaultTimeoutTick           = 20
	defaultTxBufSize             = 1024 * 1024
	defaultMaxRoutines           = 20
	defaultInitBlockSeqCap       = 50
	defaultInitPendingEntriesCap = 10
	defaultTickBufSize           = 128
	defaultProposedBlockBufSize  = 10
	defaultCommittedBlockBufSize = 10
	defaultPersistedBlockBufSize = 10
	defaultBlockLimitRate        = 100
	defaultTxLimitRate           = 100000
	defaultSendBufSize           = 20
	defaultRecvBufSize           = 100

	maxTimes   = 10
	maxTimesTx = 40
)

func checkCommonCfg(raft *consensus.CommonConfig) error {
	entries := []configuration.Entry{
		{"MaxTxCount", &raft.MaxTxCount, defaultMaxTxCount, 0, defaultMaxTxCount * maxTimesTx},
		{"MaxBlockSize", &raft.MaxBlockSize, defaultMaxBlockSize, 0, defaultMaxBlockSize * maxTimes},
		{"TickInterval", &raft.TickInterval, defaultTickInterval, 0, defaultTickInterval * maxTimes},
		{"TimeoutTick", &raft.TimeoutTick, defaultTimeoutTick, 0, defaultTimeoutTick * maxTimes},
		{"TxBufSize", &raft.TxBufSize, defaultTxBufSize, 0, defaultTxBufSize * maxTimes},
		{"MaxRoutines", &raft.MaxRoutines, defaultMaxRoutines, 1, defaultMaxRoutines * maxTimes},
		{"InitBlockSeqCap", &raft.InitBlockSeqCap, defaultInitBlockSeqCap, 0, defaultInitBlockSeqCap * maxTimes},
		{"InitPendingEntriesCap", &raft.InitPendingEntriesCap, defaultInitPendingEntriesCap, 0,
			defaultInitPendingEntriesCap * maxTimes},
		{"TickBufSize", &raft.TickBufSize, defaultTickBufSize, 0, defaultTickBufSize * maxTimes},
		{"ProposedBlockBufSize", &raft.ProposedBlockBufSize, defaultProposedBlockBufSize, 0,
			defaultProposedBlockBufSize * maxTimes},
		{"CommittedBlockBufSize", &raft.CommittedBlockBufSize, defaultCommittedBlockBufSize, 0,
			defaultCommittedBlockBufSize * maxTimes},
		{"PersistedBlockBufSize", &raft.PersistedBlockBufSize, defaultPersistedBlockBufSize, 0,
			defaultPersistedBlockBufSize * maxTimes},
		{"BlockLimitRate", &raft.BlockLimitRate, defaultBlockLimitRate, 0, defaultBlockLimitRate * maxTimes},
		{"TxLimitRate", &raft.TxLimitRate, defaultTxLimitRate, 1, defaultTxLimitRate * maxTimes},
		{"SendBufSize", &raft.SendBufSize, defaultSendBufSize, 1, defaultSendBufSize * maxTimes},
		{"RecvBufSize", &raft.RecvBufSize, defaultRecvBufSize, 1, defaultRecvBufSize * maxTimes},
	}
	list := configuration.List{Name: "Raft", Entries: entries}

	if err := checkCfgList(&list); err != nil {
		return err
	}
	return nil
}

const (
	defaultHeartbeat     = 5
	defaultElection      = 50
	defaultUnreachable   = 20
	defaultSnapThreshold = 20
	defaultMaxBatchSize  = 5
)

func checkInitStateCfg(state *consensus.InitialState) error {
	entries := []configuration.Entry{
		{"Tick", &state.Tick, defaultTickInterval, 0, defaultTickInterval * maxTimes},
		{"Heartbeat", &state.Heartbeat, defaultHeartbeat, 0, defaultHeartbeat * maxTimes},
		{"Election", &state.Election, defaultElection, 0, defaultElection * maxTimes},
		{"Unreachable", &state.Unreachable, defaultUnreachable, 0, defaultUnreachable * maxTimes},
		{"SnapThreshold", &state.SnapThreshold, defaultSnapThreshold, 0, defaultSnapThreshold * maxTimes},
		{"MaxBatchSize", &state.MaxBatchSize, defaultMaxBatchSize, 0, defaultMaxBatchSize * maxTimes},
	}
	list := configuration.List{Name: "RaftInitState", Entries: entries}

	if err := checkCfgList(&list); err != nil {
		return err
	}
	return nil
}

const (
	defaultMaxFaultNodes = 1
	defaultReqTimeout    = 2000
)

func checkFlicTypeCfg(flic *consensus.FlicConfig) error {
	entries := []configuration.Entry{
		{"MaxFaultNodes", &flic.MaxFaultNodes, defaultMaxFaultNodes, 0, defaultMaxFaultNodes * maxTimes},
		{"ReqTimeout", &flic.ReqTimeout, defaultReqTimeout, 0, defaultReqTimeout * maxTimes},
	}
	list := configuration.List{Name: "Flic", Entries: entries}

	if err := checkCfgList(&list); err != nil {
		return err
	}
	return nil
}

const (
	defaultViewTimeout    = 20
	maxViewTimeout        = 6000 // 6000 * 100ms = 10min
	defaultMaxPrunedBs    = 100
	defaultMaxBlockBatch  = 20
	defaultWaitTxDuration = 200
	maxWaitTxDuration     = 600000 // 10min
)

func checkHotstuffCfg(h *consensus.HotStuffConfig) error {
	entries := []configuration.Entry{
		{"Tick", &h.Tick, defaultTickInterval, 0, defaultTickInterval * maxTimes},
		{"ViewTimeout", &h.ViewTimeout, defaultViewTimeout, 0, maxViewTimeout},
		{"MaxPrunedBlocks", &h.MaxPrunedBlocks, defaultMaxPrunedBs, 0, defaultMaxPrunedBs * maxTimes},
		{"MaxBlockBatch", &h.MaxBlockBatch, defaultMaxBlockBatch, 0, defaultMaxPrunedBs * maxTimes},
		{"WaitTxDuration", &h.WaitTxDuration, defaultWaitTxDuration, 0, maxWaitTxDuration},
	}
	list := configuration.List{Name: "HotstuffConfig", Entries: entries}

	if err := checkCfgList(&list); err != nil {
		return err
	}
	return nil
}
