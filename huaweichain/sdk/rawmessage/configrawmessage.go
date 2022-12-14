/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package rawmessage

import (
	goerrors "errors"
	"strings"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/consensus"
	"git.huawei.com/huaweichain/sdk/config"
)

const updateHandler = "update"
const maxCertLen = 4096

// ConfigRawMessage is the definition of ConfigRawMessage.
type ConfigRawMessage struct {
	ZoneConfig      *ZoneConfig
	DomainConfig    *DomainConfig
	ConsensusConfig *ConsensusConfig
	UpdateConfig    *UpdateConfig
	RecordConfig    *RecordCert
}

// NewConfigRawMessage is used to create an instance of config raw message by specifying an message builder.
func NewConfigRawMessage(builder MsgBuilder) *ConfigRawMessage {
	return &ConfigRawMessage{
		ZoneConfig:      &ZoneConfig{builder: builder},
		DomainConfig:    &DomainConfig{builder: builder},
		ConsensusConfig: &ConsensusConfig{builder: builder},
		UpdateConfig:    &UpdateConfig{builder: builder},
		RecordConfig:    &RecordCert{builder: builder},
	}
}

// ZoneConfig is the definition of zone config.
type ZoneConfig struct {
	builder MsgBuilder
}

// BuildAddRawMsg is used to build add raw message.
func (c *ZoneConfig) BuildAddRawMsg(chainID string, zones []*common.ZoneProperties) (*TxRawMsg, error) {
	return c.BuildAddRawMsgWithTarget(chainID, zones, nil)
}

// BuildAddRawMsgWithTarget is used to build add raw message with propose target.
func (c *ZoneConfig) BuildAddRawMsgWithTarget(chainID string, zones []*common.ZoneProperties,
	target *common.ProposeTarget) (*TxRawMsg, error) {
	add := &common.ZoneOp_Add{Zones: zones}
	op := &common.ZoneOp{Op: &common.ZoneOp_Add_{Add: add}}
	tx, err := c.buildTx(chainID, op)
	if err != nil {
		return nil, errors.WithMessage(err, "build tx error")
	}
	return c.builder.GetTxRawMsgWithTarget(tx, target)
}

// BuildRemoveRawMsg is used to build remove raw message.
func (c *ZoneConfig) BuildRemoveRawMsg(chainID string, zones []string) (*TxRawMsg, error) {
	return c.BuildRemoveRawMsgWithTarget(chainID, zones, nil)
}

// BuildRemoveRawMsgWithTarget is used to build remove raw message with propose target.
func (c *ZoneConfig) BuildRemoveRawMsgWithTarget(chainID string, zones []string,
	target *common.ProposeTarget) (*TxRawMsg, error) {
	remove := &common.ZoneOp_Remove{Zones: zones}
	op := &common.ZoneOp{Op: &common.ZoneOp_Remove_{Remove: remove}}
	tx, err := c.buildTx(chainID, op)
	if err != nil {
		return nil, errors.WithMessage(err, "build tx error")
	}
	return c.builder.GetTxRawMsgWithTarget(tx, target)
}

func (c *ZoneConfig) buildTx(chainID string, op *common.ZoneOp) (*common.Transaction, error) {
	if err := c.checkOpCfg(op); err != nil {
		return nil, err
	}
	net := common.NetUpdate{
		Update: &common.NetUpdate_Zone{
			Zone: op,
		}}
	return buildNetUpdateTx(chainID, &net, c.builder)
}

func (c *ZoneConfig) checkOpCfg(op *common.ZoneOp) error {
	switch op.GetOp().(type) {
	case *common.ZoneOp_Remove_:
		return checkRemoveCfg(op.GetRemove())
	case *common.ZoneOp_Add_:
		return checkAddCfg(op.GetAdd())
	default:
		return errors.New("unknown type")
	}
}

func checkAddCfg(cfg *common.ZoneOp_Add) error {
	if len(cfg.Zones) == 0 {
		return errors.New("ZoneOp zones is empty")
	}
	zones := make([]*common.Zone, len(cfg.Zones))
	for i, z := range cfg.Zones {
		zone, err := config.BuildZone(z)
		if err != nil {
			return err
		}
		zones[i] = zone
	}
	if err := config.CheckZones(zones); err != nil {
		return err
	}
	return nil
}

func checkRemoveCfg(cfg *common.ZoneOp_Remove) error {
	if len(cfg.Zones) == 0 {
		return errors.New("ZoneOp zones is empty")
	}
	for i := 0; i < len(cfg.Zones); i++ {
		if err := config.CheckZoneID(cfg.Zones[i]); err != nil {
			return err
		}
	}
	return nil
}

// DomainConfig is the definition of domain config.
type DomainConfig struct {
	builder MsgBuilder
}

// BuildAddRawMsg is used to build add raw message.
func (c *DomainConfig) BuildAddRawMsg(chainID string, domains []string) (*TxRawMsg, error) {
	return c.BuildAddRawMsgWithTarget(chainID, domains, nil)
}

// BuildAddRawMsgWithTarget is used to build add raw message with propose target.
func (c *DomainConfig) BuildAddRawMsgWithTarget(chainID string, domains []string,
	target *common.ProposeTarget) (*TxRawMsg, error) {
	if err := config.CheckDomainPath(domains); err != nil {
		return nil, err
	}
	tx, err := c.buildAddTx(chainID, domains)
	if err != nil {
		return nil, errors.WithMessage(err, "build add tx error")
	}
	return c.builder.GetTxRawMsgWithTarget(tx, target)
}

// BuildRemoveRawMsg is used to build remove raw message.
func (c *DomainConfig) BuildRemoveRawMsg(chainID string, domains []string) (*TxRawMsg, error) {
	return c.BuildRemoveRawMsgWithTarget(chainID, domains, nil)
}

// BuildRemoveRawMsgWithTarget is used to build remove raw message with propose target.
func (c *DomainConfig) BuildRemoveRawMsgWithTarget(chainID string, domains []string,
	target *common.ProposeTarget) (*TxRawMsg, error) {
	if err := config.CheckDomainPath(domains); err != nil {
		return nil, err
	}
	tx, err := c.buildRemoveTx(chainID, domains)
	if err != nil {
		return nil, errors.WithMessage(err, "build remove tx error")
	}
	return c.builder.GetTxRawMsgWithTarget(tx, target)
}

func (c *DomainConfig) buildAddTx(chainID string, domains []string) (*common.Transaction, error) {
	return c.buildTx(chainID, domains, common.ADD)
}

func (c *DomainConfig) buildRemoveTx(chainID string, domains []string) (*common.Transaction, error) {
	return c.buildTx(chainID, domains, common.REMOVE)
}

func (c *DomainConfig) buildTx(chainID string, domains []string,
	op common.DomainOp_Op) (*common.Transaction, error) {
	net := common.NetUpdate{
		Update: &common.NetUpdate_Domain{
			Domain: &common.DomainOp{
				Op:      op,
				Domains: domains,
			}}}
	return buildNetUpdateTx(chainID, &net, c.builder)
}

func buildNetUpdateTx(chainID string, net *common.NetUpdate, builder MsgBuilder) (*common.Transaction, error) {
	set := common.ConfigSet{Value: &common.ConfigSet_Net{Net: net}}
	setBytes, err := proto.Marshal(&set)
	if err != nil {
		return nil, errors.WithMessage(err, "config set marshal error")
	}
	voteTxData := &common.VoteTxData{Payload: setBytes, Handler: updateHandler}
	return builder.BuildVoteTx(chainID, voteTxData)
}

// ConsensusConfig has param and builder for consensus config
type ConsensusConfig struct {
	param   *param
	builder MsgBuilder
}

type param struct {
	chainID  string
	nodeName string
	nodeType string
	voteType string
	opType   string
	path     string
	decrypt  config.DecryptFunc
}

const (
	voteType        = "vote"
	completeType    = "complete"
	voteHandler     = "consensus_vote"
	completeHandler = "consensus_complete"
	voter           = "voter"
	learner         = "learner"
)

// operation type const, Add, Update, remove
const (
	AddOp    = "add"
	UpdateOp = "update"
	RemoveOp = "remove"
)

var (
	errOpTypeNotSupported   = goerrors.New("consensus op type not supported")
	errNodeTypeNotSupported = goerrors.New("consensus node type not supported")
	errVoteTypeNotSupported = goerrors.New("consensus vote type not supported")
)

// BuildConfChangeRawMessage implements consensus config msg build process
func (c *ConsensusConfig) BuildConfChangeRawMessage(chainID string, nodeName string,
	opType string, nodeType string, voteType string, path string, decrypts ...config.DecryptFunc) (*TxRawMsg, error) {
	decrypt := getDecrypt(decrypts...)
	c.param = &param{
		chainID:  chainID,
		nodeName: nodeName,
		opType:   opType,
		nodeType: nodeType,
		voteType: voteType,
		path:     path,
		decrypt:  decrypt,
	}
	err := c.param.check()
	if err != nil {
		return nil, err
	}
	tx, err := c.buildConfChangeTx()
	if err != nil {
		return nil, errors.WithMessage(err, "build add tx error")
	}
	return c.builder.GetTxRawMsg(tx)
}

// UpdateConfig has builder for chain config update
type UpdateConfig struct {
	builder MsgBuilder
}

// BuildUpdateMinPltfVersionRawMessage build update config policy rawMessage
func (u *UpdateConfig) BuildUpdateMinPltfVersionRawMessage(chainID string,
	minPltfVersion string) (*TxRawMsg, error) {
	configSet := &common.ConfigSet{
		Value: &common.ConfigSet_MinPlatformVersion{MinPlatformVersion: minPltfVersion},
	}
	return u.buildUpdateChain(chainID, configSet)
}

// BuildUpdateConfPolicyRawMessage build update config policy rawMessage
func (u *UpdateConfig) BuildUpdateConfPolicyRawMessage(chainID string, policy string) (*TxRawMsg, error) {
	configSet := &common.ConfigSet{
		Value: &common.ConfigSet_ConfigPolicy{ConfigPolicy: policy},
	}
	return u.buildUpdateChain(chainID, configSet)
}

// BuildUpdateLifecycleRawMessage build update lifecycle rawMessage
func (u *UpdateConfig) BuildUpdateLifecycleRawMessage(chainID string, policy string) (*TxRawMsg, error) {
	configSet := &common.ConfigSet{
		Value: &common.ConfigSet_LifecyclePolicy{LifecyclePolicy: policy},
	}
	return u.buildUpdateChain(chainID, configSet)
}

// BuildUpdateOrgRawMessage build update org rawMessage
func (u *UpdateConfig) BuildUpdateOrgRawMessage(chainID string,
	orgUpdates *common.ConfigSet_OrgUpdates) (*TxRawMsg, error) {
	configSet := &common.ConfigSet{
		Value: &common.ConfigSet_OrgUpdates_{OrgUpdates: orgUpdates},
	}
	return u.buildUpdateChain(chainID, configSet)
}

// BuildUpdateOrgRawMessageWithYaml build update org rawMessage with yaml
func (u *UpdateConfig) BuildUpdateOrgRawMessageWithYaml(chainID string, path string,
	decrypt config.DecryptFunc) (*TxRawMsg, error) {
	orgUpdates, err := config.GetOrgUpdates(path, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get org updates error")
	}
	configSet := &common.ConfigSet{
		Value: &common.ConfigSet_OrgUpdates_{OrgUpdates: orgUpdates},
	}
	return u.buildUpdateChain(chainID, configSet)
}

// BuildChangeCertStatusRawMessage build raw message for change cert status request
func (u *UpdateConfig) BuildChangeCertStatusRawMessage(chainID string, pemCert []byte,
	action string) (*TxRawMsg, error) {
	err := checkChainID(chainID)
	if err != nil {
		return nil, errors.WithMessage(err, "the chainID is not correct")
	}

	certLen := len(pemCert)
	if certLen == 0 || certLen > maxCertLen {
		return nil, errors.New("the pem cert length is not correct")
	}

	var certStatus common.CertStatus
	actionLower := strings.ToLower(action)
	switch actionLower {
	case revocation:
		certStatus = common.CERT_REVOCATION
	case freeze:
		certStatus = common.CERT_FREEZE
	case unfreeze:
		certStatus = common.CERT_NORMAL
	default:
		return nil, errors.Errorf("can not execute %s action for this certificate", action)
	}

	cs := &common.CertWithStatus{
		Cert:   pemCert,
		Status: certStatus,
	}
	csBytes, err := proto.Marshal(cs)
	if err != nil {
		return nil, err
	}

	voteTxData := &common.VoteTxData{
		Handler: common.CERT_STATUS_CHANGE.String(),
		Payload: csBytes,
	}
	tx, err := u.builder.BuildVoteTx(chainID, voteTxData)
	if err != nil {
		return nil, err
	}
	return u.builder.GetTxRawMsg(tx)
}

func (u *UpdateConfig) buildUpdateChain(chainID string, configSet proto.Marshaler) (*TxRawMsg, error) {
	bytes, marshalErr := proto.Marshal(configSet)
	if marshalErr != nil {
		return nil, errors.WithMessage(marshalErr, "marshal ConfigSet error")
	}
	voteTxData := &common.VoteTxData{
		Handler: updateHandler,
		Payload: bytes,
	}
	transaction, txErr := u.builder.BuildVoteTx(chainID, voteTxData)
	if txErr != nil {
		return nil, errors.WithMessage(txErr, "build add tx error")
	}
	return u.builder.GetTxRawMsg(transaction)
}

func getDecrypt(decrypts ...config.DecryptFunc) config.DecryptFunc {
	decrypt := func(bytes []byte) ([]byte, error) {
		return bytes, nil
	}
	if len(decrypts) > 0 {
		decrypt = decrypts[0]
	}
	return decrypt
}

func (c *ConsensusConfig) buildConfChangeTx() (*common.Transaction, error) {
	cs, err := c.getConsenter()
	if err != nil {
		return nil, errors.WithMessage(err, "get consensus config error")
	}
	op, err := c.getConfOp()
	if err != nil {
		return nil, errors.WithMessage(err, "get config op error")
	}
	msg := &consensus.ConfigChange{Op: op, Consenter: cs}
	bytes, err := msg.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "consensus config marshal error")
	}
	voteTxData := &common.VoteTxData{Payload: bytes, Handler: c.voteTxHandler()}
	return c.builder.BuildVoteTx(c.param.chainID, voteTxData)
}

func (c *ConsensusConfig) getConsenter() (*consensus.Consenter, error) {
	// UpdateOp and RemoveOp only need the name of consenter
	if c.param.opType == UpdateOp || c.param.opType == RemoveOp {
		return &consensus.Consenter{Name: c.param.nodeName}, nil
	}
	// addOp need config information of consenter from file
	return config.NewConsensusConfig(c.param.nodeName, c.param.path, c.param.decrypt)
}

func (p *param) check() error {
	if p.nodeType != voter && p.nodeType != learner {
		return errNodeTypeNotSupported
	}
	if p.voteType != voteType && p.voteType != completeType {
		return errVoteTypeNotSupported
	}
	if p.opType != AddOp && p.opType != UpdateOp && p.opType != RemoveOp {
		return errOpTypeNotSupported
	}
	return nil
}

func (c *ConsensusConfig) voteTxHandler() string {
	if c.param.voteType == voteType {
		return voteHandler
	}
	return completeHandler
}

func (c *ConsensusConfig) getConfOp() (consensus.ConfOp, error) {
	switch c.param.opType {
	case AddOp:
		return c.getAddOp(), nil
	case UpdateOp:
		return c.getUpdateOp(), nil
	case RemoveOp:
		return consensus.Remove, nil
	default:
		return 0, errOpTypeNotSupported
	}
}

func (c *ConsensusConfig) getAddOp() consensus.ConfOp {
	if c.param.nodeType == voter {
		return consensus.AddVoter
	}
	return consensus.AddLearner
}

func (c *ConsensusConfig) getUpdateOp() consensus.ConfOp {
	if c.param.nodeType == voter {
		return consensus.UpdateToVoter
	}
	return consensus.UpdateToLearner
}

// RecordCert has builder for record cert.
type RecordCert struct {
	builder MsgBuilder
}

// BuildRecordCertRawMessage build record cert vote rawMessage with cert.
func (r *RecordCert) BuildRecordCertRawMessage(chainID string, cert []byte) (*TxRawMsg, error) {
	voteTxData := &common.VoteTxData{
		Handler: common.RECORD.String(),
		Payload: cert,
	}
	transaction, txErr := r.builder.BuildVoteTx(chainID, voteTxData)
	if txErr != nil {
		return nil, errors.WithMessage(txErr, "build add tx error")
	}
	return r.builder.GetTxRawMsg(transaction)
}
