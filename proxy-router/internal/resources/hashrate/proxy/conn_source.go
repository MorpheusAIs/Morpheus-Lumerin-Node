package proxy

import (
	"context"
	"time"

	globalInterfaces "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// ConnSource is a miner connection, a wrapper around StratumConnection
// that adds miner specific state variables
type ConnSource struct {
	// state
	userName string

	extraNonce     string // last relevant extraNonce (from subscribe or set_extranonce)
	extraNonceSize int

	versionRollingMask        string // original supported rolling mask from the miner
	versionRollingMinBitCount int    // originally sent from the miner
	currentVersionRollingMask string // current rolling mask after negotiation with server

	stats *SourceStats

	// deps
	log  globalInterfaces.ILogger
	conn *StratumConnection
}

func NewSourceConn(conn *StratumConnection, log globalInterfaces.ILogger) *ConnSource {
	return &ConnSource{
		conn:  conn,
		stats: &SourceStats{},
		log:   log,
	}
}

func (c *ConnSource) GetID() string {
	return c.conn.GetID()
}

func (c *ConnSource) Read(ctx context.Context) (interfaces.MiningMessageGeneric, error) {
	//TODO: message validation
	msg, err := c.conn.Read(ctx)
	if err != nil {
		return nil, lib.WrapError(ErrSource, err)
	}
	return msg, nil
}

func (c *ConnSource) Write(ctx context.Context, msg interfaces.MiningMessageGeneric) error {
	//TODO: message validation
	err := c.conn.Write(ctx, msg)
	if err != nil {
		return lib.WrapError(ErrSource, err)
	}
	return nil
}

func (c *ConnSource) GetExtraNonce() (extraNonce string, extraNonceSize int) {
	return c.extraNonce, c.extraNonceSize
}

func (c *ConnSource) SetExtraNonce(extraNonce string, extraNonceSize int) {
	c.extraNonce, c.extraNonceSize = extraNonce, extraNonceSize
}

func (c *ConnSource) SetVersionRolling(mask string, minBitCount int) {
	c.versionRollingMask, c.versionRollingMinBitCount = mask, minBitCount
}

func (c *ConnSource) GetVersionRolling() (mask string, minBitCount int) {
	return c.versionRollingMask, c.versionRollingMinBitCount
}

// GetNegotiatedVersionRollingMask returns actual version rolling mask after negotiation with server
func (c *ConnSource) GetNegotiatedVersionRollingMask() string {
	return c.versionRollingMask
}

// SetNegotiatedVersionRollingMask sets actual version rolling mask after negotiation with server
func (c *ConnSource) SetNegotiatedVersionRollingMask(mask string) {
	c.versionRollingMask = mask
}

func (c *ConnSource) SetUserName(userName string) {
	c.userName = userName
	c.log = c.log.Named(userName)
	c.conn.log = c.log
}

func (c *ConnSource) GetUserName() string {
	return c.userName
}

func (c *ConnSource) GetConnectedAt() time.Time {
	return c.conn.GetConnectedAt()
}

func (c *ConnSource) GetStats() *SourceStats {
	return c.stats
}
