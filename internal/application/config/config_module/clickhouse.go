package config_module

type Clickhouse struct {
	User             string
	Password         string
	IP               []string
	Dbname           string
	Size             int
	SaveTime         int
	Debug            bool
	MaxExecutionTime int
	DialTimeout      int
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  int
	BlockBufferSize  uint8
}

func NewClickhouseConfig() *Clickhouse {
	return &Clickhouse{}
}

func (c Clickhouse) GetUser() string           { return c.User }
func (c Clickhouse) GetPassword() string       { return c.Password }
func (c Clickhouse) GetIP() []string           { return c.IP }
func (c Clickhouse) GetDbname() string         { return c.Dbname }
func (c Clickhouse) GetSize() int              { return c.Size }
func (c Clickhouse) GetSaveTime() int          { return c.SaveTime }
func (c Clickhouse) GetDebug() bool            { return c.Debug }
func (c Clickhouse) GetMaxExecutionTime() int  { return c.MaxExecutionTime }
func (c Clickhouse) GetDialTimeout() int       { return c.DialTimeout }
func (c Clickhouse) GetMaxOpenConns() int      { return c.MaxOpenConns }
func (c Clickhouse) GetMaxIdleConns() int      { return c.MaxIdleConns }
func (c Clickhouse) GetConnMaxLifetime() int   { return c.ConnMaxLifetime }
func (c Clickhouse) GetBlockBufferSize() uint8 { return c.BlockBufferSize }
